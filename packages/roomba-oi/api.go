// Package roombaoi defines command payloads for the iRobot Roomba 600 Open
// Interface (OI).
package roombaoi

import (
	"fmt"
	"io"
	"reflect"
)

// Opcode identifies an OI command byte.
type Opcode uint8

const (
	OpcodeReset             Opcode = 7
	OpcodeStart             Opcode = 128
	OpcodeBaud              Opcode = 129
	OpcodeControl           Opcode = 130
	OpcodeSafe              Opcode = 131
	OpcodeFull              Opcode = 132
	OpcodePower             Opcode = 133
	OpcodeSpot              Opcode = 134
	OpcodeClean             Opcode = 135
	OpcodeMax               Opcode = 136
	OpcodeDrive             Opcode = 137
	OpcodeMotors            Opcode = 138
	OpcodeLEDs              Opcode = 139
	OpcodeSong              Opcode = 140
	OpcodePlay              Opcode = 141
	OpcodeSensors           Opcode = 142
	OpcodeSeekDock          Opcode = 143
	OpcodePWMMotors         Opcode = 144
	OpcodeDriveDirect       Opcode = 145
	OpcodeDrivePWM          Opcode = 146
	OpcodeStream            Opcode = 148
	OpcodeQueryList         Opcode = 149
	OpcodePauseResumeStream Opcode = 150
	OpcodeSchedulingLEDs    Opcode = 162
	OpcodeDigitLEDsRaw      Opcode = 163
	OpcodeDigitLEDsASCII    Opcode = 164
	OpcodeButtons           Opcode = 165
	OpcodeSchedule          Opcode = 167
	OpcodeSetDayTime        Opcode = 168
	OpcodeStop              Opcode = 173
)

// Command is implemented by every OI command type.
type Command interface {
	Opcode() Opcode
	io.WriterTo
}

func writeCommand(w io.Writer, opcode Opcode, data []byte) (int64, error) {
	bytes := make([]byte, 1+len(data))
	bytes[0] = byte(opcode)
	copy(bytes[1:], data)

	return writeAll(w, bytes)
}

func writeAll(w io.Writer, bytes []byte) (n int64, err error) {
	for len(bytes) > 0 {
		written, writeErr := w.Write(bytes)
		n += int64(written)
		bytes = bytes[written:]
		if writeErr != nil {
			return n, writeErr
		}
		if written == 0 {
			return n, io.ErrShortWrite
		}
	}

	return n, nil
}

func int16Bytes(value int16) [2]byte {
	unsigned := uint16(value)
	return [2]byte{byte(unsigned >> 8), byte(unsigned)}
}

// StartCommand starts the OI and enters Passive mode.
type StartCommand struct{}

func (StartCommand) Opcode() Opcode { return OpcodeStart }

func (command StartCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), nil)
}

// ResetCommand resets the robot and exits the OI.
type ResetCommand struct{}

func (ResetCommand) Opcode() Opcode { return OpcodeReset }

func (command ResetCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), nil)
}

// StopCommand stops the OI and exits the OI.
type StopCommand struct{}

func (StopCommand) Opcode() Opcode { return OpcodeStop }

func (command StopCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), nil)
}

// BaudCode is one of the baud-rate codes accepted by BaudCommand (0 through 11).
type BaudCode uint8

// BaudCommand changes the OI serial baud rate. The host must wait 100 ms
// before communicating at the new rate.
type BaudCommand struct {
	Code BaudCode
}

func (BaudCommand) Opcode() Opcode { return OpcodeBaud }

func (command BaudCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), []byte{byte(command.Code)})
}

// ControlCommand is the legacy alias of SafeCommand using opcode 130.
type ControlCommand struct{}

func (ControlCommand) Opcode() Opcode { return OpcodeControl }

func (command ControlCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), nil)
}

// SafeCommand enters Safe mode.
type SafeCommand struct{}

func (SafeCommand) Opcode() Opcode { return OpcodeSafe }

func (command SafeCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), nil)
}

// FullCommand enters Full mode.
type FullCommand struct{}

func (FullCommand) Opcode() Opcode { return OpcodeFull }

func (command FullCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), nil)
}

// CleanCommand starts or pauses the default cleaning cycle.
type CleanCommand struct{}

func (CleanCommand) Opcode() Opcode { return OpcodeClean }

func (command CleanCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), nil)
}

// MaxCommand starts or pauses the Max cleaning cycle.
type MaxCommand struct{}

func (MaxCommand) Opcode() Opcode { return OpcodeMax }

func (command MaxCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), nil)
}

// SpotCommand starts or pauses the Spot cleaning cycle.
type SpotCommand struct{}

func (SpotCommand) Opcode() Opcode { return OpcodeSpot }

func (command SpotCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), nil)
}

// SeekDockCommand directs the robot to seek its charging dock.
type SeekDockCommand struct{}

func (SeekDockCommand) Opcode() Opcode { return OpcodeSeekDock }

func (command SeekDockCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), nil)
}

// PowerCommand powers down the robot.
type PowerCommand struct{}

func (PowerCommand) Opcode() Opcode { return OpcodePower }

func (command PowerCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), nil)
}

// Weekdays is a bitset of scheduled-cleaning days.
type Weekdays uint8

const (
	WeekdaySunday    Weekdays = 1 << 0
	WeekdayMonday    Weekdays = 1 << 1
	WeekdayTuesday   Weekdays = 1 << 2
	WeekdayWednesday Weekdays = 1 << 3
	WeekdayThursday  Weekdays = 1 << 4
	WeekdayFriday    Weekdays = 1 << 5
	WeekdaySaturday  Weekdays = 1 << 6
)

// Day identifies a day of the week for SetDayTimeCommand.
type Day uint8

const (
	Sunday Day = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

// ClockTime is a 24-hour time used by the OI clock and cleaning schedule.
type ClockTime struct {
	Hour   uint8 // 0 through 23.
	Minute uint8 // 0 through 59.
}

// ScheduleCommand sets the weekly cleaning schedule. Times are ordered Sunday
// through Saturday. A zero Days value and zero Times disable scheduled cleaning.
type ScheduleCommand struct {
	Days  Weekdays
	Times [7]ClockTime
}

func (ScheduleCommand) Opcode() Opcode { return OpcodeSchedule }

func (command ScheduleCommand) WriteTo(w io.Writer) (int64, error) {
	data := make([]byte, 0, 15)
	data = append(data, byte(command.Days))
	for _, time := range command.Times {
		data = append(data, time.Hour, time.Minute)
	}
	return writeCommand(w, command.Opcode(), data)
}

// SetDayTimeCommand sets the OI clock.
type SetDayTimeCommand struct {
	Day  Day
	Time ClockTime
}

func (SetDayTimeCommand) Opcode() Opcode { return OpcodeSetDayTime }

func (command SetDayTimeCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), []byte{
		byte(command.Day), command.Time.Hour, command.Time.Minute,
	})
}

// DriveRadius is the turning radius for DriveCommand, in millimeters.
type DriveRadius int16

const (
	DriveStraight             DriveRadius = -32768
	DriveStraightAlternative  DriveRadius = 32767
	DriveTurnClockwise        DriveRadius = -1
	DriveTurnCounterclockwise DriveRadius = 1
)

// DriveCommand controls both wheels using a shared velocity and turning radius.
type DriveCommand struct {
	VelocityMMPerSec int16       // -500 through 500.
	RadiusMM         DriveRadius // -2000 through 2000, or a Drive* special value.
}

func (DriveCommand) Opcode() Opcode { return OpcodeDrive }

func (command DriveCommand) WriteTo(w io.Writer) (int64, error) {
	velocity := int16Bytes(command.VelocityMMPerSec)
	radius := int16Bytes(int16(command.RadiusMM))
	return writeCommand(w, command.Opcode(), []byte{
		velocity[0], velocity[1], radius[0], radius[1],
	})
}

// DriveDirectCommand independently controls the right and left wheel velocities.
type DriveDirectCommand struct {
	RightVelocityMMPerSec int16 // -500 through 500.
	LeftVelocityMMPerSec  int16 // -500 through 500.
}

func (DriveDirectCommand) Opcode() Opcode { return OpcodeDriveDirect }

func (command DriveDirectCommand) WriteTo(w io.Writer) (int64, error) {
	right := int16Bytes(command.RightVelocityMMPerSec)
	left := int16Bytes(command.LeftVelocityMMPerSec)
	return writeCommand(w, command.Opcode(), []byte{
		right[0], right[1], left[0], left[1],
	})
}

// DrivePWMCommand independently controls raw wheel PWM values.
type DrivePWMCommand struct {
	RightPWM int16 // -255 through 255.
	LeftPWM  int16 // -255 through 255.
}

func (DrivePWMCommand) Opcode() Opcode { return OpcodeDrivePWM }

func (command DrivePWMCommand) WriteTo(w io.Writer) (int64, error) {
	right := int16Bytes(command.RightPWM)
	left := int16Bytes(command.LeftPWM)
	return writeCommand(w, command.Opcode(), []byte{
		right[0], right[1], left[0], left[1],
	})
}

// MotorFlags controls the brush and vacuum motors.
type MotorFlags uint8

const (
	MotorSideBrush        MotorFlags = 1 << 0
	MotorVacuum           MotorFlags = 1 << 1
	MotorMainBrush        MotorFlags = 1 << 2
	MotorSideBrushReverse MotorFlags = 1 << 3
	MotorMainBrushReverse MotorFlags = 1 << 4
)

// MotorsCommand enables the selected motors at full speed. Reverse flags affect
// the matching brush only when that brush is enabled.
type MotorsCommand struct {
	Motors MotorFlags
}

func (MotorsCommand) Opcode() Opcode { return OpcodeMotors }

func (command MotorsCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), []byte{byte(command.Motors)})
}

// PWMMotorsCommand sets individual brush and vacuum motor duty cycles.
type PWMMotorsCommand struct {
	MainBrushPWM int8  // -127 through 127.
	SideBrushPWM int8  // -127 through 127.
	VacuumPWM    uint8 // 0 through 127.
}

func (PWMMotorsCommand) Opcode() Opcode { return OpcodePWMMotors }

func (command PWMMotorsCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), []byte{
		byte(command.MainBrushPWM), byte(command.SideBrushPWM), command.VacuumPWM,
	})
}

// LEDFlags controls the common Roomba 600 LEDs.
type LEDFlags uint8

const (
	LEDDebris     LEDFlags = 1 << 0
	LEDSpot       LEDFlags = 1 << 1
	LEDDock       LEDFlags = 1 << 2
	LEDCheckRobot LEDFlags = 1 << 3
)

// LEDsCommand controls the common LEDs and the bicolor power LED.
type LEDsCommand struct {
	LEDs           LEDFlags
	PowerColor     uint8 // 0 is green; 255 is red.
	PowerIntensity uint8 // 0 is off; 255 is full intensity.
}

func (LEDsCommand) Opcode() Opcode { return OpcodeLEDs }

func (command LEDsCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), []byte{
		byte(command.LEDs), command.PowerColor, command.PowerIntensity,
	})
}

// SchedulingLEDsCommand controls the weekday and scheduling LEDs found on
// Roomba 560 and 570 models.
type SchedulingLEDsCommand struct {
	WeekdayLEDs    uint8
	SchedulingLEDs uint8
}

func (SchedulingLEDsCommand) Opcode() Opcode { return OpcodeSchedulingLEDs }

func (command SchedulingLEDsCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), []byte{
		command.WeekdayLEDs, command.SchedulingLEDs,
	})
}

// DigitLEDsRawCommand controls the four seven-segment displays on Roomba 560
// and 570. Digits are ordered left-to-right as 3, 2, 1, 0.
type DigitLEDsRawCommand struct {
	Digits [4]uint8
}

func (DigitLEDsRawCommand) Opcode() Opcode { return OpcodeDigitLEDsRaw }

func (command DigitLEDsRawCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), command.Digits[:])
}

// DigitLEDsASCIICommand writes ASCII characters to the four seven-segment
// displays on Roomba 560 and 570. Each byte must be in the range 32 through 126.
type DigitLEDsASCIICommand struct {
	Digits [4]byte
}

func (DigitLEDsASCIICommand) Opcode() Opcode { return OpcodeDigitLEDsASCII }

func (command DigitLEDsASCIICommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), command.Digits[:])
}

// ButtonFlags identifies the buttons to press with ButtonsCommand.
type ButtonFlags uint8

const (
	ButtonClock    ButtonFlags = 1 << 0
	ButtonSchedule ButtonFlags = 1 << 1
	ButtonDay      ButtonFlags = 1 << 2
	ButtonHour     ButtonFlags = 1 << 3
	ButtonMinute   ButtonFlags = 1 << 4
	ButtonDock     ButtonFlags = 1 << 5
	ButtonSpot     ButtonFlags = 1 << 6
	ButtonClean    ButtonFlags = 1 << 7
)

// ButtonsCommand presses the selected buttons for one-sixth of a second.
type ButtonsCommand struct {
	Buttons ButtonFlags
}

func (ButtonsCommand) Opcode() Opcode { return OpcodeButtons }

func (command ButtonsCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), []byte{byte(command.Buttons)})
}

// SongNumber identifies one of the OI song slots (0 through 4).
type SongNumber uint8

// SongNote is a MIDI note and its duration in 1/64-second increments. A note
// outside 31 through 127 is treated by the OI as a rest.
type SongNote struct {
	Number   uint8
	Duration uint8
}

// SongCommand stores one to sixteen notes in a song slot.
type SongCommand struct {
	Number SongNumber
	Notes  []SongNote
}

func (SongCommand) Opcode() Opcode { return OpcodeSong }

func (command SongCommand) WriteTo(w io.Writer) (int64, error) {
	if len(command.Notes) < 1 || len(command.Notes) > 16 {
		return 0, fmt.Errorf("roomba OI song length %d is outside the allowed range 1 through 16", len(command.Notes))
	}

	data := make([]byte, 0, 2+2*len(command.Notes))
	data = append(data, byte(command.Number), byte(len(command.Notes)))
	for _, note := range command.Notes {
		data = append(data, note.Number, note.Duration)
	}
	return writeCommand(w, command.Opcode(), data)
}

// PlayCommand plays the selected song slot.
type PlayCommand struct {
	Number SongNumber
}

func (PlayCommand) Opcode() Opcode { return OpcodePlay }

func (command PlayCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), []byte{byte(command.Number)})
}

// SensorPacketID identifies a sensor packet requested from the OI.
type SensorPacketID uint8

const (
	SensorPacketBumpsAndWheelDrops         SensorPacketID = 7
	SensorPacketWall                       SensorPacketID = 8
	SensorPacketCliffLeft                  SensorPacketID = 9
	SensorPacketCliffFrontLeft             SensorPacketID = 10
	SensorPacketCliffFrontRight            SensorPacketID = 11
	SensorPacketCliffRight                 SensorPacketID = 12
	SensorPacketVirtualWall                SensorPacketID = 13
	SensorPacketWheelOvercurrents          SensorPacketID = 14
	SensorPacketDirtDetect                 SensorPacketID = 15
	SensorPacketUnused                     SensorPacketID = 16
	SensorPacketInfraredOmni               SensorPacketID = 17
	SensorPacketButtons                    SensorPacketID = 18
	SensorPacketDistance                   SensorPacketID = 19
	SensorPacketAngle                      SensorPacketID = 20
	SensorPacketChargingState              SensorPacketID = 21
	SensorPacketVoltage                    SensorPacketID = 22
	SensorPacketCurrent                    SensorPacketID = 23
	SensorPacketTemperature                SensorPacketID = 24
	SensorPacketBatteryCharge              SensorPacketID = 25
	SensorPacketBatteryCapacity            SensorPacketID = 26
	SensorPacketWallSignal                 SensorPacketID = 27
	SensorPacketCliffLeftSignal            SensorPacketID = 28
	SensorPacketCliffFrontLeftSignal       SensorPacketID = 29
	SensorPacketCliffFrontRightSignal      SensorPacketID = 30
	SensorPacketCliffRightSignal           SensorPacketID = 31
	SensorPacketUnused32                   SensorPacketID = 32
	SensorPacketUnused33                   SensorPacketID = 33
	SensorPacketChargingSources            SensorPacketID = 34
	SensorPacketOIMode                     SensorPacketID = 35
	SensorPacketSongNumber                 SensorPacketID = 36
	SensorPacketSongPlaying                SensorPacketID = 37
	SensorPacketStreamPacketCount          SensorPacketID = 38
	SensorPacketRequestedVelocity          SensorPacketID = 39
	SensorPacketRequestedRadius            SensorPacketID = 40
	SensorPacketRequestedRightVelocity     SensorPacketID = 41
	SensorPacketRequestedLeftVelocity      SensorPacketID = 42
	SensorPacketLeftEncoderCounts          SensorPacketID = 43
	SensorPacketRightEncoderCounts         SensorPacketID = 44
	SensorPacketLightBumper                SensorPacketID = 45
	SensorPacketLightBumpLeftSignal        SensorPacketID = 46
	SensorPacketLightBumpFrontLeftSignal   SensorPacketID = 47
	SensorPacketLightBumpCenterLeftSignal  SensorPacketID = 48
	SensorPacketLightBumpCenterRightSignal SensorPacketID = 49
	SensorPacketLightBumpFrontRightSignal  SensorPacketID = 50
	SensorPacketLightBumpRightSignal       SensorPacketID = 51
	SensorPacketInfraredLeft               SensorPacketID = 52
	SensorPacketInfraredRight              SensorPacketID = 53
	SensorPacketLeftMotorCurrent           SensorPacketID = 54
	SensorPacketRightMotorCurrent          SensorPacketID = 55
	SensorPacketMainBrushMotorCurrent      SensorPacketID = 56
	SensorPacketSideBrushMotorCurrent      SensorPacketID = 57
	SensorPacketStasis                     SensorPacketID = 58
)

// SensorPacket is implemented by a decoded OI sensor packet.
// If `T` implements `SensorPacket`, `*T` should implement `sensorPacketPtr`.
type SensorPacket interface {
	ID() SensorPacketID
}

type sensorPacketPtr interface {
	io.ReaderFrom
}

var sensorPacketIDByType = map[reflect.Type]SensorPacketID{
	reflect.TypeFor[BumpsAndWheelDropsPacket]():         SensorPacketBumpsAndWheelDrops,
	reflect.TypeFor[WallPacket]():                       SensorPacketWall,
	reflect.TypeFor[CliffLeftPacket]():                  SensorPacketCliffLeft,
	reflect.TypeFor[CliffFrontLeftPacket]():             SensorPacketCliffFrontLeft,
	reflect.TypeFor[CliffFrontRightPacket]():            SensorPacketCliffFrontRight,
	reflect.TypeFor[CliffRightPacket]():                 SensorPacketCliffRight,
	reflect.TypeFor[VirtualWallPacket]():                SensorPacketVirtualWall,
	reflect.TypeFor[WheelOvercurrentsPacket]():          SensorPacketWheelOvercurrents,
	reflect.TypeFor[DirtDetectPacket]():                 SensorPacketDirtDetect,
	reflect.TypeFor[UnusedPacket]():                     SensorPacketUnused,
	reflect.TypeFor[InfraredOmniPacket]():               SensorPacketInfraredOmni,
	reflect.TypeFor[ButtonsSensorPacket]():              SensorPacketButtons,
	reflect.TypeFor[DistancePacket]():                   SensorPacketDistance,
	reflect.TypeFor[AnglePacket]():                      SensorPacketAngle,
	reflect.TypeFor[ChargingStatePacket]():              SensorPacketChargingState,
	reflect.TypeFor[VoltagePacket]():                    SensorPacketVoltage,
	reflect.TypeFor[CurrentPacket]():                    SensorPacketCurrent,
	reflect.TypeFor[TemperaturePacket]():                SensorPacketTemperature,
	reflect.TypeFor[BatteryChargePacket]():              SensorPacketBatteryCharge,
	reflect.TypeFor[BatteryCapacityPacket]():            SensorPacketBatteryCapacity,
	reflect.TypeFor[WallSignalPacket]():                 SensorPacketWallSignal,
	reflect.TypeFor[CliffLeftSignalPacket]():            SensorPacketCliffLeftSignal,
	reflect.TypeFor[CliffFrontLeftSignalPacket]():       SensorPacketCliffFrontLeftSignal,
	reflect.TypeFor[CliffFrontRightSignalPacket]():      SensorPacketCliffFrontRightSignal,
	reflect.TypeFor[CliffRightSignalPacket]():           SensorPacketCliffRightSignal,
	reflect.TypeFor[Unused32Packet]():                   SensorPacketUnused32,
	reflect.TypeFor[Unused33Packet]():                   SensorPacketUnused33,
	reflect.TypeFor[ChargingSourcesPacket]():            SensorPacketChargingSources,
	reflect.TypeFor[OIModePacket]():                     SensorPacketOIMode,
	reflect.TypeFor[SongNumberPacket]():                 SensorPacketSongNumber,
	reflect.TypeFor[SongPlayingPacket]():                SensorPacketSongPlaying,
	reflect.TypeFor[StreamPacketCountPacket]():          SensorPacketStreamPacketCount,
	reflect.TypeFor[RequestedVelocityPacket]():          SensorPacketRequestedVelocity,
	reflect.TypeFor[RequestedRadiusPacket]():            SensorPacketRequestedRadius,
	reflect.TypeFor[RequestedRightVelocityPacket]():     SensorPacketRequestedRightVelocity,
	reflect.TypeFor[RequestedLeftVelocityPacket]():      SensorPacketRequestedLeftVelocity,
	reflect.TypeFor[LeftEncoderCountsPacket]():          SensorPacketLeftEncoderCounts,
	reflect.TypeFor[RightEncoderCountsPacket]():         SensorPacketRightEncoderCounts,
	reflect.TypeFor[LightBumperPacket]():                SensorPacketLightBumper,
	reflect.TypeFor[LightBumpLeftSignalPacket]():        SensorPacketLightBumpLeftSignal,
	reflect.TypeFor[LightBumpFrontLeftSignalPacket]():   SensorPacketLightBumpFrontLeftSignal,
	reflect.TypeFor[LightBumpCenterLeftSignalPacket]():  SensorPacketLightBumpCenterLeftSignal,
	reflect.TypeFor[LightBumpCenterRightSignalPacket](): SensorPacketLightBumpCenterRightSignal,
	reflect.TypeFor[LightBumpFrontRightSignalPacket]():  SensorPacketLightBumpFrontRightSignal,
	reflect.TypeFor[LightBumpRightSignalPacket]():       SensorPacketLightBumpRightSignal,
	reflect.TypeFor[InfraredLeftPacket]():               SensorPacketInfraredLeft,
	reflect.TypeFor[InfraredRightPacket]():              SensorPacketInfraredRight,
	reflect.TypeFor[LeftMotorCurrentPacket]():           SensorPacketLeftMotorCurrent,
	reflect.TypeFor[RightMotorCurrentPacket]():          SensorPacketRightMotorCurrent,
	reflect.TypeFor[MainBrushMotorCurrentPacket]():      SensorPacketMainBrushMotorCurrent,
	reflect.TypeFor[SideBrushMotorCurrentPacket]():      SensorPacketSideBrushMotorCurrent,
	reflect.TypeFor[StasisPacket]():                     SensorPacketStasis,
}

// UInt8SensorPacket is the shared decoder for one-byte unsigned sensor values.
type UInt8SensorPacket struct {
	Value uint8
}

func (packet *UInt8SensorPacket) ReadFrom(reader io.Reader) (int64, error) {
	var data [1]byte
	n, err := io.ReadFull(reader, data[:])
	packet.Value = data[0]
	return int64(n), err
}

// Int8SensorPacket is the shared decoder for one-byte signed sensor values.
type Int8SensorPacket struct {
	Value int8
}

func (packet *Int8SensorPacket) ReadFrom(reader io.Reader) (int64, error) {
	var data [1]byte
	n, err := io.ReadFull(reader, data[:])
	packet.Value = int8(data[0])
	return int64(n), err
}

// UInt16SensorPacket is the shared decoder for two-byte unsigned sensor values.
type UInt16SensorPacket struct {
	Value uint16
}

func (packet *UInt16SensorPacket) ReadFrom(reader io.Reader) (int64, error) {
	var data [2]byte
	n, err := io.ReadFull(reader, data[:])
	packet.Value = uint16(data[0])<<8 | uint16(data[1])
	return int64(n), err
}

// Int16SensorPacket is the shared decoder for two-byte signed sensor values.
type Int16SensorPacket struct {
	Value int16
}

func (packet *Int16SensorPacket) ReadFrom(reader io.Reader) (int64, error) {
	var data [2]byte
	n, err := io.ReadFull(reader, data[:])
	packet.Value = int16(uint16(data[0])<<8 | uint16(data[1]))
	return int64(n), err
}

// BooleanSensorPacket is the shared decoder for one-byte OI boolean values.
type BooleanSensorPacket struct {
	Value bool
}

func (packet *BooleanSensorPacket) ReadFrom(reader io.Reader) (int64, error) {
	var data [1]byte
	n, err := io.ReadFull(reader, data[:])
	packet.Value = data[0] != 0
	return int64(n), err
}

type BumpsAndWheelDropsPacket struct{ UInt8SensorPacket }

func (BumpsAndWheelDropsPacket) ID() SensorPacketID { return SensorPacketBumpsAndWheelDrops }

type WallPacket struct{ BooleanSensorPacket }

func (WallPacket) ID() SensorPacketID { return SensorPacketWall }

type CliffLeftPacket struct{ BooleanSensorPacket }

func (CliffLeftPacket) ID() SensorPacketID { return SensorPacketCliffLeft }

type CliffFrontLeftPacket struct{ BooleanSensorPacket }

func (CliffFrontLeftPacket) ID() SensorPacketID { return SensorPacketCliffFrontLeft }

type CliffFrontRightPacket struct{ BooleanSensorPacket }

func (CliffFrontRightPacket) ID() SensorPacketID { return SensorPacketCliffFrontRight }

type CliffRightPacket struct{ BooleanSensorPacket }

func (CliffRightPacket) ID() SensorPacketID { return SensorPacketCliffRight }

type VirtualWallPacket struct{ BooleanSensorPacket }

func (VirtualWallPacket) ID() SensorPacketID { return SensorPacketVirtualWall }

type WheelOvercurrentsPacket struct{ UInt8SensorPacket }

func (WheelOvercurrentsPacket) ID() SensorPacketID { return SensorPacketWheelOvercurrents }

type DirtDetectPacket struct{ UInt8SensorPacket }

func (DirtDetectPacket) ID() SensorPacketID { return SensorPacketDirtDetect }

type UnusedPacket struct{ UInt8SensorPacket }

func (UnusedPacket) ID() SensorPacketID { return SensorPacketUnused }

type InfraredOmniPacket struct{ UInt8SensorPacket }

func (InfraredOmniPacket) ID() SensorPacketID { return SensorPacketInfraredOmni }

type ButtonsSensorPacket struct{ UInt8SensorPacket }

func (ButtonsSensorPacket) ID() SensorPacketID { return SensorPacketButtons }

type DistancePacket struct{ Int16SensorPacket }

func (DistancePacket) ID() SensorPacketID { return SensorPacketDistance }

type AnglePacket struct{ Int16SensorPacket }

func (AnglePacket) ID() SensorPacketID { return SensorPacketAngle }

type ChargingStatePacket struct{ UInt8SensorPacket }

func (ChargingStatePacket) ID() SensorPacketID { return SensorPacketChargingState }

type VoltagePacket struct{ UInt16SensorPacket }

func (VoltagePacket) ID() SensorPacketID { return SensorPacketVoltage }

type CurrentPacket struct{ Int16SensorPacket }

func (CurrentPacket) ID() SensorPacketID { return SensorPacketCurrent }

type TemperaturePacket struct{ Int8SensorPacket }

func (TemperaturePacket) ID() SensorPacketID { return SensorPacketTemperature }

type BatteryChargePacket struct{ UInt16SensorPacket }

func (BatteryChargePacket) ID() SensorPacketID { return SensorPacketBatteryCharge }

type BatteryCapacityPacket struct{ UInt16SensorPacket }

func (BatteryCapacityPacket) ID() SensorPacketID { return SensorPacketBatteryCapacity }

type WallSignalPacket struct{ UInt16SensorPacket }

func (WallSignalPacket) ID() SensorPacketID { return SensorPacketWallSignal }

type CliffLeftSignalPacket struct{ UInt16SensorPacket }

func (CliffLeftSignalPacket) ID() SensorPacketID { return SensorPacketCliffLeftSignal }

type CliffFrontLeftSignalPacket struct{ UInt16SensorPacket }

func (CliffFrontLeftSignalPacket) ID() SensorPacketID { return SensorPacketCliffFrontLeftSignal }

type CliffFrontRightSignalPacket struct{ UInt16SensorPacket }

func (CliffFrontRightSignalPacket) ID() SensorPacketID { return SensorPacketCliffFrontRightSignal }

type CliffRightSignalPacket struct{ UInt16SensorPacket }

func (CliffRightSignalPacket) ID() SensorPacketID { return SensorPacketCliffRightSignal }

type Unused32Packet struct{ UInt8SensorPacket }

func (Unused32Packet) ID() SensorPacketID { return SensorPacketUnused32 }

type Unused33Packet struct{ UInt16SensorPacket }

func (Unused33Packet) ID() SensorPacketID { return SensorPacketUnused33 }

type ChargingSourcesPacket struct{ UInt8SensorPacket }

func (ChargingSourcesPacket) ID() SensorPacketID { return SensorPacketChargingSources }

type OIModePacket struct{ UInt8SensorPacket }

func (OIModePacket) ID() SensorPacketID { return SensorPacketOIMode }

type SongNumberPacket struct{ UInt8SensorPacket }

func (SongNumberPacket) ID() SensorPacketID { return SensorPacketSongNumber }

type SongPlayingPacket struct{ BooleanSensorPacket }

func (SongPlayingPacket) ID() SensorPacketID { return SensorPacketSongPlaying }

type StreamPacketCountPacket struct{ UInt8SensorPacket }

func (StreamPacketCountPacket) ID() SensorPacketID { return SensorPacketStreamPacketCount }

type RequestedVelocityPacket struct{ Int16SensorPacket }

func (RequestedVelocityPacket) ID() SensorPacketID { return SensorPacketRequestedVelocity }

type RequestedRadiusPacket struct{ Int16SensorPacket }

func (RequestedRadiusPacket) ID() SensorPacketID { return SensorPacketRequestedRadius }

type RequestedRightVelocityPacket struct{ Int16SensorPacket }

func (RequestedRightVelocityPacket) ID() SensorPacketID { return SensorPacketRequestedRightVelocity }

type RequestedLeftVelocityPacket struct{ Int16SensorPacket }

func (RequestedLeftVelocityPacket) ID() SensorPacketID { return SensorPacketRequestedLeftVelocity }

type LeftEncoderCountsPacket struct{ Int16SensorPacket }

func (LeftEncoderCountsPacket) ID() SensorPacketID { return SensorPacketLeftEncoderCounts }

type RightEncoderCountsPacket struct{ Int16SensorPacket }

func (RightEncoderCountsPacket) ID() SensorPacketID { return SensorPacketRightEncoderCounts }

type LightBumperPacket struct{ UInt8SensorPacket }

func (LightBumperPacket) ID() SensorPacketID { return SensorPacketLightBumper }

type LightBumpLeftSignalPacket struct{ UInt16SensorPacket }

func (LightBumpLeftSignalPacket) ID() SensorPacketID { return SensorPacketLightBumpLeftSignal }

type LightBumpFrontLeftSignalPacket struct{ UInt16SensorPacket }

func (LightBumpFrontLeftSignalPacket) ID() SensorPacketID {
	return SensorPacketLightBumpFrontLeftSignal
}

type LightBumpCenterLeftSignalPacket struct{ UInt16SensorPacket }

func (LightBumpCenterLeftSignalPacket) ID() SensorPacketID {
	return SensorPacketLightBumpCenterLeftSignal
}

type LightBumpCenterRightSignalPacket struct{ UInt16SensorPacket }

func (LightBumpCenterRightSignalPacket) ID() SensorPacketID {
	return SensorPacketLightBumpCenterRightSignal
}

type LightBumpFrontRightSignalPacket struct{ UInt16SensorPacket }

func (LightBumpFrontRightSignalPacket) ID() SensorPacketID {
	return SensorPacketLightBumpFrontRightSignal
}

type LightBumpRightSignalPacket struct{ UInt16SensorPacket }

func (LightBumpRightSignalPacket) ID() SensorPacketID { return SensorPacketLightBumpRightSignal }

type InfraredLeftPacket struct{ UInt8SensorPacket }

func (InfraredLeftPacket) ID() SensorPacketID { return SensorPacketInfraredLeft }

type InfraredRightPacket struct{ UInt8SensorPacket }

func (InfraredRightPacket) ID() SensorPacketID { return SensorPacketInfraredRight }

type LeftMotorCurrentPacket struct{ Int16SensorPacket }

func (LeftMotorCurrentPacket) ID() SensorPacketID { return SensorPacketLeftMotorCurrent }

type RightMotorCurrentPacket struct{ Int16SensorPacket }

func (RightMotorCurrentPacket) ID() SensorPacketID { return SensorPacketRightMotorCurrent }

type MainBrushMotorCurrentPacket struct{ Int16SensorPacket }

func (MainBrushMotorCurrentPacket) ID() SensorPacketID { return SensorPacketMainBrushMotorCurrent }

type SideBrushMotorCurrentPacket struct{ Int16SensorPacket }

func (SideBrushMotorCurrentPacket) ID() SensorPacketID { return SensorPacketSideBrushMotorCurrent }

type StasisPacket struct{ UInt8SensorPacket }

func (StasisPacket) ID() SensorPacketID { return SensorPacketStasis }

// SensorCommand is an alias for SensorsCommand, which requests one sensor packet.
type SensorCommand = SensorsCommand

// NewSensorCommand creates a command that requests a packet of type T.
func NewSensorCommand[T SensorPacket]() SensorCommand {
	id, err := sensorPacketID(reflect.TypeFor[T]())
	if err != nil {
		panic(err)
	}
	return SensorCommand{PacketID: id}
}

// NewQueryListCommand creates a query-list command for the SensorPacket fields
// of struct type T, in declaration order. Fields may store packets inline or as
// concrete packet pointers.
func NewQueryListCommand[T any]() QueryListCommand {
	ids, err := sensorPacketIDs(reflect.TypeFor[T]())
	if err != nil {
		panic(err)
	}
	return QueryListCommand{PacketIDs: ids}
}

// NewStreamCommand creates a stream command for the SensorPacket fields of
// struct type T, in declaration order. See NewQueryListCommand for requirements.
func NewStreamCommand[T any]() StreamCommand {
	ids, err := sensorPacketIDs(reflect.TypeFor[T]())
	if err != nil {
		panic(err)
	}
	return StreamCommand{PacketIDs: ids}
}

// ReadSensorPackets reads one packet into each SensorPacket field of a new
// value of struct type T, in declaration order. Nil pointer fields are
// initialized before decoding.
func ReadSensorPackets(sequence any, reader io.Reader) (n int64, err error) {
	value := reflect.ValueOf(sequence).Elem()
	if value.Kind() != reflect.Struct {
		return 0, fmt.Errorf("sensor packet sequence type must be a struct, got %s", value.Type())
	}

	for index := 0; index < value.NumField(); index++ {
		field := value.Field(index).Addr()
		packet, ok := field.Interface().(sensorPacketPtr)
		if !ok {
			return 0, fmt.Errorf("sensor packet field %d does not implement sensorPacketPtr", index)
		}
		read, readErr := packet.ReadFrom(reader)
		n += read
		if readErr != nil {
			return n, readErr
		}
	}

	return n, nil
}

func sensorPacketID(typ reflect.Type) (SensorPacketID, error) {
	zeroVal, valid := reflect.Zero(typ).Interface().(SensorPacket)
	if !valid {
		return 0, fmt.Errorf("sensor packet struct %s must implement SensorPacket", typ)
	}
	return zeroVal.ID(), nil
}

func sensorPacketIDs(typeOfSequence reflect.Type) ([]SensorPacketID, error) {
	if typeOfSequence.Kind() != reflect.Struct {
		return nil, fmt.Errorf("sensor packet sequence type must be a struct, got %s", typeOfSequence)
	}

	// TODO: Is it more efficient to use reflect.Zero on the whole struct?
	packetIDs := make([]SensorPacketID, 0, typeOfSequence.NumField())
	for index := 0; index < typeOfSequence.NumField(); index++ {
		fieldType := typeOfSequence.Field(index).Type
		id, err := sensorPacketID(fieldType)
		if err != nil {
			return nil, err
		}
		packetIDs = append(packetIDs, id)
	}
	return packetIDs, nil
}

// SensorsCommand requests one sensor packet.
type SensorsCommand struct {
	PacketID SensorPacketID
}

func (SensorsCommand) Opcode() Opcode { return OpcodeSensors }

func (command SensorsCommand) WriteTo(w io.Writer) (int64, error) {
	return writeCommand(w, command.Opcode(), []byte{byte(command.PacketID)})
}

// QueryListCommand requests each sensor packet once, in the specified order.
type QueryListCommand struct {
	PacketIDs []SensorPacketID
}

func (QueryListCommand) Opcode() Opcode { return OpcodeQueryList }

func (command QueryListCommand) WriteTo(w io.Writer) (int64, error) {
	return writePacketList(w, command.Opcode(), command.PacketIDs)
}

// StreamCommand starts a recurring stream of the specified sensor packets.
type StreamCommand struct {
	PacketIDs []SensorPacketID
}

func (StreamCommand) Opcode() Opcode { return OpcodeStream }

func (command StreamCommand) WriteTo(w io.Writer) (int64, error) {
	return writePacketList(w, command.Opcode(), command.PacketIDs)
}

// PauseResumeStreamCommand stops or resumes the most recently requested stream.
type PauseResumeStreamCommand struct {
	Resume bool
}

func (PauseResumeStreamCommand) Opcode() Opcode { return OpcodePauseResumeStream }

func (command PauseResumeStreamCommand) WriteTo(w io.Writer) (int64, error) {
	state := byte(0)
	if command.Resume {
		state = 1
	}
	return writeCommand(w, command.Opcode(), []byte{state})
}

func writePacketList(w io.Writer, opcode Opcode, packetIDs []SensorPacketID) (int64, error) {
	if len(packetIDs) > 255 {
		return 0, fmt.Errorf("roomba OI packet list has %d packets; maximum is 255", len(packetIDs))
	}

	data := make([]byte, 1+len(packetIDs))
	data[0] = byte(len(packetIDs))
	for index, packetID := range packetIDs {
		data[index+1] = byte(packetID)
	}
	return writeCommand(w, opcode, data)
}
