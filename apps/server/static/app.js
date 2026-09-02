const boardElement = document.getElementById('board');
const robotBoardElement = document.getElementById('robot-board');
const newGameButton = document.getElementById('new-game-button');
const cancelMoveButton = document.getElementById('cancel-move-button');
const messageLabel = document.getElementById('message-label');

let currentBoardState = null;

// Robot state
let currentRobots = new Map(); // robotId -> RobotInfo
let boardConfig = { boardSize: 8, captureColumns: 0, cellSizeMM: 450, robotDiameterMM: 340 };

// A move in progress. `selectedPieceId` is set as soon as a piece is picked,
// and stays set until the move is explicitly cancelled or accepted - clicking
// another piece or an empty square does *not* clear it. `movePath` accumulates
// the squares clicked so far (starting with the piece's own square), and
// `candidateMoves` is narrowed down to the legal moves for the selected piece
// that are still consistent with `movePath`, so it always reflects the
// possible remaining continuations (relevant for multi-jump sequences).
let selectedPieceId = null;
let movePath = [];
let candidateMoves = [];

function squareKey(row, col) {
  return `${row},${col}`;
}

// Label for an actual checkers square, e.g. "a1". Only meaningful for
// columns within the board itself (not the capture columns beside it).
function squareLabel(row, col) {
  return `${String.fromCharCode(97 + col)}${row + 1}`;
}

function clearSelection() {
  selectedPieceId = null;
  movePath = [];
  candidateMoves = [];
}

// Legal squares the active piece could move to next, given the squares
// already clicked in `movePath`.
function nextStepSquareKeys() {
  const keys = new Set();
  for (const move of candidateMoves) {
    if (Array.isArray(move.path) && move.path.length > movePath.length) {
      const next = move.path[movePath.length];
      keys.add(squareKey(next.row, next.col));
    }
  }
  return keys;
}

// The candidate move that exactly matches the squares clicked so far, if any.
// Once this exists, the in-progress move is a complete, submittable move.
function getCompletedMove() {
  return candidateMoves.find((move) => Array.isArray(move.path) && move.path.length === movePath.length) ?? null;
}

// Squares of pieces jumped over so far in the in-progress move, so they can
// be hidden from the board even though the server hasn't captured them yet.
function jumpedSquareKeysInPath() {
  const keys = new Set();
  for (let i = 0; i < movePath.length - 1; i++) {
    const from = movePath[i];
    const to = movePath[i + 1];
    if (Math.abs(to.row - from.row) === 2) {
      keys.add(squareKey((from.row + to.row) / 2, (from.col + to.col) / 2));
    }
  }
  return keys;
}

function updateMoveControls() {
  if (cancelMoveButton) {
    cancelMoveButton.disabled = !selectedPieceId;
  }
}

function renderBoard() {
  if (!boardElement || !messageLabel || !currentBoardState) {
    return;
  }

  const pieces = Array.isArray(currentBoardState.pieces) ? currentBoardState.pieces : [];
  const boardSize = Number.isInteger(currentBoardState.boardSize) && currentBoardState.boardSize > 0 ? currentBoardState.boardSize : 8;
  const captureColumns = Number.isInteger(currentBoardState.captureColumns) && currentBoardState.captureColumns >= 0
    ? currentBoardState.captureColumns
    : 0;
  const totalCols = boardSize + 2 * captureColumns;

  // Pieces (on-board or captured) all carry a Row/Col position resolved by
  // the server, on the same extended grid rendered below.
  const piecesByPosition = new Map(pieces.map((piece) => [squareKey(piece.row, piece.col), piece]));

  const pathSquareKeys = new Set(movePath.map((position) => squareKey(position.row, position.col)));
  const currentSquareKey = movePath.length > 0 ? squareKey(movePath[movePath.length - 1].row, movePath[movePath.length - 1].col) : null;
  const legalNextSquareKeys = selectedPieceId ? nextStepSquareKeys() : new Set();
  const hasMoved = movePath.length > 1;
  const originSquareKey = movePath.length > 0 ? squareKey(movePath[0].row, movePath[0].col) : null;
  const jumpedSquareKeys = hasMoved ? jumpedSquareKeysInPath() : new Set();
  const activePiece = hasMoved && selectedPieceId ? pieces.find((piece) => piece.id === selectedPieceId) ?? null : null;
  // Only ring pieces when no piece is selected yet.
  const moveablePieceIds = selectedPieceId
    ? new Set()
    : new Set(Object.keys(currentBoardState.legalMovesByPiece ?? {}));

  if (currentBoardState.gameOver && typeof currentBoardState.gameOver.winner === 'string') {
    messageLabel.textContent = `${currentBoardState.gameOver.winner} wins!`;
  } else if (typeof currentBoardState.turn === 'string' && currentBoardState.turn.length > 0) {
    messageLabel.textContent = `${currentBoardState.turn} to move.`;
  } else {
    messageLabel.textContent = 'Start a new game';
  }

  boardElement.replaceChildren();
  boardElement.style.gridTemplateColumns = `repeat(${totalCols}, 1fr)`;
  boardElement.style.gridTemplateRows = `repeat(${boardSize}, 1fr)`;
  boardElement.style.aspectRatio = `${totalCols} / ${boardSize}`;

  for (let row = boardSize - 1; row >= 0; row--) {
    for (let col = -captureColumns; col < boardSize + captureColumns; col++) {
      const key = squareKey(row, col);
      const isBoardSquare = col >= 0 && col < boardSize;
      const square = document.createElement('div');
      square.dataset.row = String(row);
      square.dataset.col = String(col);

      if (isBoardSquare) {
        const isDark = (row + col) % 2 === 0;
        square.className = `square ${isDark ? 'square--dark' : 'square--light'}`;
        square.setAttribute('role', 'gridcell');
        square.setAttribute('aria-label', `Square ${squareLabel(row, col)}`);
      } else {
        square.className = 'square square--capture';
        square.setAttribute('aria-label', 'Capture area');
      }

      // Reflect the in-progress move: the active piece visually sits at the
      // last clicked square rather than its server-side origin, and any piece
      // it jumped over so far is hidden even though it isn't captured yet.
      let piece = piecesByPosition.get(key) ?? null;
      if (piece && jumpedSquareKeys.has(key)) {
        piece = null;
      } else if (piece && hasMoved && key === originSquareKey) {
        piece = null;
      } else if (!piece && hasMoved && key === currentSquareKey && activePiece) {
        piece = activePiece;
      }

      if (piece) {
        square.dataset.pieceId = piece.id;
      }

      if (key !== currentSquareKey && pathSquareKeys.has(key)) {
        square.classList.add('square--path');
      }

      if (legalNextSquareKeys.has(key)) {
        square.classList.add('square--legal-move');
      }

      if (piece) {
        const pieceElement = document.createElement('div');
        pieceElement.className = typeof piece.classes === 'string' ? piece.classes : 'piece';
        if (key === currentSquareKey) {
          pieceElement.classList.add('piece--selected');
        } else if (moveablePieceIds.has(piece.id)) {
          pieceElement.classList.add('piece--moveable');
        }
        pieceElement.setAttribute('aria-label', `${piece.side} ${piece.kind}`);
        square.appendChild(pieceElement);
      }

      boardElement.appendChild(square);
    }
  }

  updateMoveControls();
}

function selectPiece(pieceId, row, col) {
  const legalMovesByPiece = currentBoardState && typeof currentBoardState === 'object'
    ? currentBoardState.legalMovesByPiece
    : null;
  const moves = legalMovesByPiece && Array.isArray(legalMovesByPiece[pieceId])
    ? legalMovesByPiece[pieceId]
    : [];

  // A piece with no legal moves can't start a move sequence, so it stays
  // deselectable by clicking elsewhere on the board.
  if (moves.length === 0) {
    return;
  }

  selectedPieceId = pieceId;
  movePath = [{ row, col }];
  candidateMoves = moves;
  renderBoard();
}

function extendMovePath(row, col) {
  const clickedKey = squareKey(row, col);
  if (!nextStepSquareKeys().has(clickedKey)) {
    return;
  }

  const stepIndex = movePath.length;
  movePath = [...movePath, { row, col }];
  candidateMoves = candidateMoves.filter((move) => {
    return Array.isArray(move.path)
      && move.path.length > stepIndex
      && squareKey(move.path[stepIndex].row, move.path[stepIndex].col) === clickedKey;
  });
  renderBoard();

  // Auto-accept as soon as the clicked-through path exactly matches a legal
  // move - there's no separate confirmation step.
  const completedMove = getCompletedMove();
  if (completedMove) {
    void submitCompletedMove(completedMove);
  }
}

async function submitCompletedMove(move) {
  if (cancelMoveButton) {
    cancelMoveButton.disabled = true;
  }

  try {
    const response = await fetch('/api/moves', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path: move.path })
    });

    if (response.ok) {
      clearSelection();
      // Board state will be updated via the game.updated SSE event.
    } else {
      const body = await response.json().catch(() => null);
      const message = body?.error?.message ?? 'Move rejected';
      console.error('Move rejected:', message);
      updateMoveControls();
    }
  } catch (error) {
    console.error('Failed to submit move:', error);
    updateMoveControls();
  }
}

if (boardElement) {
  boardElement.addEventListener('click', (event) => {
    const square = event.target instanceof Element ? event.target.closest('.square') : null;
    if (!(square instanceof HTMLElement)) {
      return;
    }

    const row = Number.parseInt(square.dataset.row ?? '', 10);
    const col = Number.parseInt(square.dataset.col ?? '', 10);
    if (!Number.isInteger(row) || !Number.isInteger(col)) {
      return;
    }

    const clickedKey = squareKey(row, col);
    const pieceId = square.dataset.pieceId;

    if (selectedPieceId) {
      if (nextStepSquareKeys().has(clickedKey)) {
        extendMovePath(row, col);
        return;
      }

      // A move sequence has been initiated (at least one step already taken):
      // clicking another piece or an empty square does nothing - the only way
      // to deselect from here on is the Cancel button.
      if (movePath.length > 1) {
        return;
      }

      // Only the piece itself has been picked so far, with no step taken yet
      // - it can still be freely deselected or swapped for another piece.
      const wasSelectedPiece = pieceId === selectedPieceId;
      clearSelection();
      if (pieceId && !wasSelectedPiece) {
        selectPiece(pieceId, row, col);
        return;
      }

      renderBoard();
      return;
    }

    if (pieceId) {
      selectPiece(pieceId, row, col);
    } else {
      cancelMove();
    }
  });
}

function cancelMove() {
  if (!selectedPieceId) {
    return;
  }

  clearSelection();
  renderBoard();
}

if (cancelMoveButton) {
  cancelMoveButton.addEventListener('click', () => {
    cancelMove();
  });
}

if (newGameButton) {
  newGameButton.addEventListener('click', async () => {
    newGameButton.disabled = true;
    try {
      const response = await fetch('/api/new-game', { method: 'POST' });
      if (!response.ok) {
        console.error('Failed to start new game:', response.status);
      }
      // Board state will be updated via the game.updated SSE event.
    } catch (error) {
      console.error('Failed to start new game:', error);
    } finally {
      newGameButton.disabled = false;
    }
  });
}

document.addEventListener('keydown', (event) => {
  if (event.key !== 'Escape') {
    return;
  }

  cancelMove();
});


// --- Robot board ---

// Returns the CSS left/top percentages and width percentage for a robot,
// relative to the robot board element (0,0 = top-left corner).
//
// The robot board is widened by capture columns on each side, just like the
// game board, so its width in mm covers more than just the actual board -
// but since those columns are added symmetrically, the board's horizontal
// center still lines up with the container's. Only horizontal scaling needs
// the wider total; rows (and so vertical scaling) are unaffected.
function robotCSSPercent(robot) {
  const boardHeightMM = boardConfig.boardSize * boardConfig.cellSizeMM;
  const totalCols = boardConfig.boardSize + 2 * boardConfig.captureColumns;
  const containerWidthMM = totalCols * boardConfig.cellSizeMM;
  const leftPct = 50 + (robot.pose.x_mm / containerWidthMM) * 100;
  // Game Y increases upward; CSS top increases downward.
  const topPct = 50 - (robot.pose.y_mm / boardHeightMM) * 100;
  const sizePct = (boardConfig.robotDiameterMM / containerWidthMM) * 100;
  return { leftPct, topPct, sizePct };
}

// Returns the `side` ('red' or 'black') of the checker piece a robot is
// currently assigned to, or null if it isn't assigned to one.
function robotPieceSide(robot) {
  const pieces = currentBoardState && Array.isArray(currentBoardState.pieces) ? currentBoardState.pieces : [];
  const piece = pieces.find((piece) => piece.id === robot.piece_id);
  return piece ? piece.side : null;
}

// Creates or repositions the DOM element for a single robot.
function syncRobotElement(robot) {
  if (!robotBoardElement) return;
  const elemId = `robot-${CSS.escape(robot.id)}`;
  let el = document.getElementById(elemId);
  if (!el) {
    el = document.createElement('div');
    el.id = elemId;
    el.textContent = 'R';
    robotBoardElement.appendChild(el);
  }
  const side = robotPieceSide(robot);
  el.className = side ? `robot robot--${side}` : 'robot';
  const { leftPct, topPct, sizePct } = robotCSSPercent(robot);
  el.style.left = `${leftPct}%`;
  el.style.top = `${topPct}%`;
  el.style.width = `${sizePct}%`;
}

// Renders the checkerboard grid of the robot board and all current robots.
//
// Mirrors the game board: extended by capture columns on each side (kept
// empty here, with a neutral background) so the two boards line up.
function renderRobotBoard() {
  if (!robotBoardElement || !currentBoardState) return;
  const boardSize = currentBoardState.boardSize || boardConfig.boardSize;
  const captureColumns = currentBoardState.captureColumns ?? boardConfig.captureColumns;
  const totalCols = boardSize + 2 * captureColumns;

  robotBoardElement.replaceChildren();
  robotBoardElement.style.gridTemplateColumns = `repeat(${totalCols}, 1fr)`;
  robotBoardElement.style.gridTemplateRows = `repeat(${boardSize}, 1fr)`;
  robotBoardElement.style.aspectRatio = `${totalCols} / ${boardSize}`;

  for (let row = boardSize - 1; row >= 0; row--) {
    for (let col = -captureColumns; col < boardSize + captureColumns; col++) {
      const square = document.createElement('div');
      const isBoardSquare = col >= 0 && col < boardSize;
      if (isBoardSquare) {
        const isDark = (row + col) % 2 === 0;
        square.className = `square ${isDark ? 'square--dark' : 'square--light'}`;
      } else {
        square.className = 'square square--capture';
      }
      robotBoardElement.appendChild(square);
    }
  }
  for (const robot of currentRobots.values()) {
    syncRobotElement(robot);
  }
}

// --- SSE ---

function loadEventStream() {
  const stream = new EventSource('/api/events');
  stream.addEventListener('error', error => {
    console.error('Event stream error', error);
  });
  // state.snapshot is sent once on initial connect with the full app state.
  stream.addEventListener('state.snapshot', event => {
    const snapshot = JSON.parse(event.data);
    currentBoardState = snapshot.game;
    if (typeof snapshot.cellSizeMM === 'number') boardConfig.cellSizeMM = snapshot.cellSizeMM;
    if (typeof snapshot.robotDiameterMM === 'number') boardConfig.robotDiameterMM = snapshot.robotDiameterMM;
    if (snapshot.game && typeof snapshot.game.boardSize === 'number') boardConfig.boardSize = snapshot.game.boardSize;
    if (snapshot.game && typeof snapshot.game.captureColumns === 'number') boardConfig.captureColumns = snapshot.game.captureColumns;
    currentRobots.clear();
    for (const robot of (snapshot.robots || [])) {
      currentRobots.set(robot.id, robot);
    }
    clearSelection();
    renderBoard();
    renderRobotBoard();
  });
  // game.updated is sent whenever the game state changes.
  stream.addEventListener('game.updated', event => {
    currentBoardState = JSON.parse(event.data);
    if (typeof currentBoardState.boardSize === 'number') boardConfig.boardSize = currentBoardState.boardSize;
    if (typeof currentBoardState.captureColumns === 'number') boardConfig.captureColumns = currentBoardState.captureColumns;
    clearSelection();
    renderBoard();
    // Re-render the checkerboard cells; robot.updated events will reposition robots.
    renderRobotBoard();
  });
  // robot.updated is sent whenever a robot's position changes.
  stream.addEventListener('robot.updated', event => {
    const robot = JSON.parse(event.data);
    currentRobots.set(robot.id, robot);
    syncRobotElement(robot);
  });
}

loadEventStream();
