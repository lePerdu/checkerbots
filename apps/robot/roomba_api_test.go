package main

import (
	"bytes"
	"errors"
	"io"
	"slices"
	"testing"
)

func TestCommandsWriteTo(t *testing.T) {
	schedule := ScheduleCommand{
		Days: WeekdayWednesday | WeekdayFriday,
		Times: [7]ClockTime{
			{}, {}, {}, {Hour: 15}, {}, {Hour: 10, Minute: 36}, {},
		},
	}

	tests := []struct {
		name    string
		command Command
		want    []byte
	}{
		{"start", StartCommand{}, []byte{128}},
		{"reset", ResetCommand{}, []byte{7}},
		{"stop", StopCommand{}, []byte{173}},
		{"baud", BaudCommand{Code: 11}, []byte{129, 11}},
		{"control", ControlCommand{}, []byte{130}},
		{"safe", SafeCommand{}, []byte{131}},
		{"full", FullCommand{}, []byte{132}},
		{"clean", CleanCommand{}, []byte{135}},
		{"max", MaxCommand{}, []byte{136}},
		{"spot", SpotCommand{}, []byte{134}},
		{"seek dock", SeekDockCommand{}, []byte{143}},
		{"power", PowerCommand{}, []byte{133}},
		{"schedule", schedule, []byte{167, 40, 0, 0, 0, 0, 0, 0, 15, 0, 0, 0, 10, 36, 0, 0}},
		{"set day time", SetDayTimeCommand{Day: Friday, Time: ClockTime{Hour: 10, Minute: 36}}, []byte{168, 5, 10, 36}},
		{"drive", DriveCommand{VelocityMMPerSec: -200, RadiusMM: 500}, []byte{137, 255, 56, 1, 244}},
		{"drive direct", DriveDirectCommand{RightVelocityMMPerSec: -500, LeftVelocityMMPerSec: 500}, []byte{145, 254, 12, 1, 244}},
		{"drive PWM", DrivePWMCommand{RightPWM: -255, LeftPWM: 255}, []byte{146, 255, 1, 0, 255}},
		{"motors", MotorsCommand{Motors: MotorMainBrush | MotorSideBrushReverse}, []byte{138, 12}},
		{"PWM motors", PWMMotorsCommand{MainBrushPWM: -127, SideBrushPWM: 127, VacuumPWM: 64}, []byte{144, 129, 127, 64}},
		{"LEDs", LEDsCommand{LEDs: LEDDock, PowerColor: 0, PowerIntensity: 128}, []byte{139, 4, 0, 128}},
		{"scheduling LEDs", SchedulingLEDsCommand{WeekdayLEDs: 1, SchedulingLEDs: 2}, []byte{162, 1, 2}},
		{"raw digits", DigitLEDsRawCommand{Digits: [4]uint8{1, 2, 3, 4}}, []byte{163, 1, 2, 3, 4}},
		{"ASCII digits", DigitLEDsASCIICommand{Digits: [4]byte{'A', 'B', 'C', 'D'}}, []byte{164, 'A', 'B', 'C', 'D'}},
		{"buttons", ButtonsCommand{Buttons: ButtonDock | ButtonClean}, []byte{165, 160}},
		{"song", SongCommand{Number: 1, Notes: []SongNote{{Number: 60, Duration: 32}, {Number: 62, Duration: 16}}}, []byte{140, 1, 2, 60, 32, 62, 16}},
		{"play", PlayCommand{Number: 1}, []byte{141, 1}},
		{"sensors", SensorsCommand{PacketID: 7}, []byte{142, 7}},
		{"query list", QueryListCommand{PacketIDs: []SensorPacketID{7, 13}}, []byte{149, 2, 7, 13}},
		{"stream", StreamCommand{PacketIDs: []SensorPacketID{29, 13}}, []byte{148, 2, 29, 13}},
		{"pause stream", PauseResumeStreamCommand{}, []byte{150, 0}},
		{"resume stream", PauseResumeStreamCommand{Resume: true}, []byte{150, 1}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var destination bytes.Buffer
			n, err := test.command.WriteTo(&destination)
			if err != nil {
				t.Fatalf("WriteTo() error = %v", err)
			}
			if n != int64(len(test.want)) {
				t.Errorf("WriteTo() wrote %d bytes, want %d", n, len(test.want))
			}
			if !bytes.Equal(destination.Bytes(), test.want) {
				t.Errorf("WriteTo() = %v, want %v", destination.Bytes(), test.want)
			}
		})
	}
}

func TestVariableLengthCommandsRejectInvalidLengths(t *testing.T) {
	var destination bytes.Buffer

	if _, err := (SongCommand{}).WriteTo(&destination); err == nil {
		t.Error("empty SongCommand.WriteTo() error = nil, want an error")
	}

	packetIDs := make([]SensorPacketID, 256)
	if _, err := (QueryListCommand{PacketIDs: packetIDs}).WriteTo(&destination); err == nil {
		t.Error("oversized QueryListCommand.WriteTo() error = nil, want an error")
	}
}

type partialWriter struct {
	limit int
	data  []byte
}

func (writer *partialWriter) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	written := min(writer.limit, len(data))
	writer.data = append(writer.data, data[:written]...)
	return written, nil
}

type noProgressWriter struct{}

func (noProgressWriter) Write([]byte) (int, error) { return 0, nil }

func TestWriteToHandlesPartialWrites(t *testing.T) {
	writer := &partialWriter{limit: 2}
	command := DriveCommand{VelocityMMPerSec: -200, RadiusMM: 500}

	n, err := command.WriteTo(writer)
	if err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	if n != 5 {
		t.Errorf("WriteTo() wrote %d bytes, want 5", n)
	}
	want := []byte{137, 255, 56, 1, 244}
	if !bytes.Equal(writer.data, want) {
		t.Errorf("WriteTo() = %v, want %v", writer.data, want)
	}

	_, err = StartCommand{}.WriteTo(noProgressWriter{})
	if !errors.Is(err, io.ErrShortWrite) {
		t.Errorf("WriteTo() error = %v, want io.ErrShortWrite", err)
	}
}

func TestSensorPacketsImplementSensorPacket(t *testing.T) {
	packets := []SensorPacket{
		&BumpsAndWheelDropsPacket{}, &WallPacket{}, &CliffLeftPacket{}, &CliffFrontLeftPacket{},
		&CliffFrontRightPacket{}, &CliffRightPacket{}, &VirtualWallPacket{}, &WheelOvercurrentsPacket{},
		&DirtDetectPacket{}, &UnusedPacket{}, &InfraredOmniPacket{}, &ButtonsSensorPacket{},
		&DistancePacket{}, &AnglePacket{}, &ChargingStatePacket{}, &VoltagePacket{}, &CurrentPacket{},
		&TemperaturePacket{}, &BatteryChargePacket{}, &BatteryCapacityPacket{}, &WallSignalPacket{},
		&CliffLeftSignalPacket{}, &CliffFrontLeftSignalPacket{}, &CliffFrontRightSignalPacket{},
		&CliffRightSignalPacket{}, &Unused32Packet{}, &Unused33Packet{}, &ChargingSourcesPacket{},
		&OIModePacket{}, &SongNumberPacket{}, &SongPlayingPacket{}, &StreamPacketCountPacket{},
		&RequestedVelocityPacket{}, &RequestedRadiusPacket{}, &RequestedRightVelocityPacket{},
		&RequestedLeftVelocityPacket{}, &LeftEncoderCountsPacket{}, &RightEncoderCountsPacket{},
		&LightBumperPacket{}, &LightBumpLeftSignalPacket{}, &LightBumpFrontLeftSignalPacket{},
		&LightBumpCenterLeftSignalPacket{}, &LightBumpCenterRightSignalPacket{},
		&LightBumpFrontRightSignalPacket{}, &LightBumpRightSignalPacket{}, &InfraredLeftPacket{},
		&InfraredRightPacket{}, &LeftMotorCurrentPacket{}, &RightMotorCurrentPacket{},
		&MainBrushMotorCurrentPacket{}, &SideBrushMotorCurrentPacket{}, &StasisPacket{},
	}

	if len(packets) != 52 {
		t.Fatalf("sensor packet count = %d, want 52", len(packets))
	}
	for _, packet := range packets {
		if packet.ID() < SensorPacketBumpsAndWheelDrops || packet.ID() > SensorPacketStasis {
			t.Errorf("unexpected sensor packet ID %d", packet.ID())
		}
	}
}

type sensorPacketSequence struct {
	Wall    WallPacket
	Voltage VoltagePacket
	Current CurrentPacket
}

func TestSensorPacketSequenceCreators(t *testing.T) {
	sensorCommand := NewSensorCommand[VoltagePacket]()
	if sensorCommand.PacketID != SensorPacketVoltage {
		t.Errorf("NewSensorCommand() packet ID = %d, want %d", sensorCommand.PacketID, SensorPacketVoltage)
	}

	query := NewQueryListCommand[sensorPacketSequence]()
	stream := NewStreamCommand[sensorPacketSequence]()
	wantIDs := []SensorPacketID{SensorPacketWall, SensorPacketVoltage, SensorPacketCurrent}
	if !slices.Equal(query.PacketIDs, wantIDs) {
		t.Errorf("NewQueryListCommand() IDs = %v, want %v", query.PacketIDs, wantIDs)
	}
	if !slices.Equal(stream.PacketIDs, wantIDs) {
		t.Errorf("NewStreamCommand() IDs = %v, want %v", stream.PacketIDs, wantIDs)
	}
}

func TestReadSensorPackets(t *testing.T) {
	var sequence sensorPacketSequence
	n, err := ReadSensorPackets(&sequence, bytes.NewReader([]byte{1, 0x12, 0x34, 0xff, 0x38}))
	if err != nil {
		t.Errorf("ReadSensorPackets() error = %v, want nil", err)
	}
	if n != 5 {
		t.Errorf("ReadSensorPackets() bytes read = %d, want 5", n)
	}
	if !sequence.Wall.Value {
		t.Errorf("ReadSensorPackets() Wall.Value = %t, want true", sequence.Wall.Value)
	}
	if sequence.Voltage.Value != 0x1234 {
		t.Errorf("ReadSensorPackets() Voltage.Value = %#x, want %#x", sequence.Voltage.Value, 0x1234)
	}
	if sequence.Current.Value != -200 {
		t.Errorf("ReadSensorPackets() Current.Value = %d, want -200", sequence.Current.Value)
	}
}

func TestSensorPacketReadFrom(t *testing.T) {
	voltage := VoltagePacket{}
	n, err := voltage.ReadFrom(bytes.NewReader([]byte{0x12, 0x34}))
	if err != nil || n != 2 || voltage.Value != 0x1234 {
		t.Errorf("VoltagePacket.ReadFrom() = (%d, %v, %#x), want (2, nil, 0x1234)", n, err, voltage.Value)
	}

	current := CurrentPacket{}
	n, err = current.ReadFrom(bytes.NewReader([]byte{0xff, 0x38}))
	if err != nil || n != 2 || current.Value != -200 {
		t.Errorf("CurrentPacket.ReadFrom() = (%d, %v, %d), want (2, nil, -200)", n, err, current.Value)
	}

	wall := WallPacket{}
	n, err = wall.ReadFrom(bytes.NewReader([]byte{1}))
	if err != nil || n != 1 || !wall.Value {
		t.Errorf("WallPacket.ReadFrom() = (%d, %v, %t), want (1, nil, true)", n, err, wall.Value)
	}

	angle := AnglePacket{}
	n, err = angle.ReadFrom(bytes.NewReader([]byte{0xff}))
	if !errors.Is(err, io.ErrUnexpectedEOF) || n != 1 {
		t.Errorf("AnglePacket.ReadFrom() = (%d, %v), want (1, io.ErrUnexpectedEOF)", n, err)
	}
}
