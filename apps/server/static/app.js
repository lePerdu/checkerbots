const boardElement = document.getElementById('board');
const newGameButton = document.getElementById('new-game-button');
const cancelMoveButton = document.getElementById('cancel-move-button');
const messageLabel = document.getElementById('message-label');

let currentBoardState = null;

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

function findCellWithPiece(boardCells, pieceId) {
  return boardCells.find((cell) => cell.piece && cell.piece.id === pieceId) ?? null;
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

  const boardCells = Array.isArray(currentBoardState.boardCells) ? currentBoardState.boardCells : [];
  const boardSize = Number.isInteger(currentBoardState.boardSize) && currentBoardState.boardSize > 0 ? currentBoardState.boardSize : 8;
  const pathSquareKeys = new Set(movePath.map((position) => squareKey(position.row, position.col)));
  const currentSquareKey = movePath.length > 0 ? squareKey(movePath[movePath.length - 1].row, movePath[movePath.length - 1].col) : null;
  const legalNextSquareKeys = selectedPieceId ? nextStepSquareKeys() : new Set();
  const hasMoved = movePath.length > 1;
  const originSquareKey = movePath.length > 0 ? squareKey(movePath[0].row, movePath[0].col) : null;
  const jumpedSquareKeys = hasMoved ? jumpedSquareKeysInPath() : new Set();
  const activePieceCell = hasMoved && selectedPieceId ? findCellWithPiece(boardCells, selectedPieceId) : null;
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
  boardElement.style.gridTemplateColumns = `repeat(${boardSize}, 1fr)`;
  boardElement.style.gridTemplateRows = `repeat(${boardSize}, 1fr)`;

  for (const cell of boardCells) {
    const key = squareKey(cell.row, cell.col);
    const square = document.createElement('div');
    square.className = `square ${cell.isDark ? 'square--dark' : 'square--light'}`;
    square.setAttribute('role', 'gridcell');
    square.setAttribute('aria-label', `Square ${cell.squareLabel}`);
    square.dataset.row = String(cell.row);
    square.dataset.col = String(cell.col);

    // Reflect the in-progress move: the active piece visually sits at the
    // last clicked square rather than its server-side origin, and any piece
    // it jumped over so far is hidden even though it isn't captured yet.
    let piece = cell.piece;
    if (piece && jumpedSquareKeys.has(key)) {
      piece = null;
    } else if (piece && hasMoved && key === originSquareKey) {
      piece = null;
    } else if (!piece && hasMoved && key === currentSquareKey && activePieceCell) {
      piece = activePieceCell.piece;
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

  updateMoveControls();
}

async function fetchJSON(url, options) {
  const response = await fetch(url, options);
  if (!response.ok) {
    throw new Error(`request failed: ${response.status}`);
  }
  return response.json();
}

async function loadBoard() {
  try {
    currentBoardState = await fetchJSON('/api/board', {
      headers: {
        Accept: 'application/json'
      }
    });
    clearSelection();
    renderBoard();
  } catch (error) {
    console.error('Failed to load board', error);
    if (messageLabel) {
      messageLabel.textContent = 'Start a new game';
    }
  }
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
    currentBoardState = await fetchJSON('/api/moves', {
      method: 'POST',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ path: move.path })
    });
    clearSelection();
    renderBoard();
  } catch (error) {
    console.error('Failed to apply move', error);
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
      currentBoardState = await fetchJSON('/games/new', {
        method: 'POST',
        headers: {
          Accept: 'application/json'
        }
      });
      clearSelection();
      renderBoard();
    } catch (error) {
      console.error('Failed to start new game', error);
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

void loadBoard();
