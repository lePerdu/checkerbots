const boardElement = document.getElementById('board');
const newGameButton = document.getElementById('new-game-button');
const turnIndicator = document.getElementById('turn-indicator');

let currentBoardState = null;
let selectedPieceId = null;
let selectedSquareKey = null;
let legalMoveSquareKeys = [];

function squareKey(row, col) {
  return `${row},${col}`;
}

function clearSelection() {
  selectedPieceId = null;
  selectedSquareKey = null;
  legalMoveSquareKeys = [];
}

function renderBoard() {
  if (!boardElement || !turnIndicator || !currentBoardState) {
    return;
  }

  const boardCells = Array.isArray(currentBoardState.boardCells) ? currentBoardState.boardCells : [];
  const boardSize = Number.isInteger(currentBoardState.boardSize) && currentBoardState.boardSize > 0 ? currentBoardState.boardSize : 8;
  const legalMoveSquareKeySet = new Set(legalMoveSquareKeys);

  turnIndicator.textContent = typeof currentBoardState.turn === 'string' ? currentBoardState.turn : 'Unknown';
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

    if (cell.piece) {
      square.dataset.pieceId = cell.piece.id;
    }

    if (key === selectedSquareKey) {
      square.classList.add('square--selected');
    }

    if (legalMoveSquareKeySet.has(key)) {
      square.classList.add('square--legal-move');
    }

    if (cell.piece) {
      const piece = document.createElement('div');
      piece.className = typeof cell.piece.classes === 'string' ? cell.piece.classes : 'piece';
      piece.setAttribute('aria-label', `${cell.piece.side} ${cell.piece.kind}`);
      piece.textContent = typeof cell.piece.label === 'string' ? cell.piece.label : '';
      piece.dataset.pieceId = cell.piece.id;
      piece.dataset.row = String(cell.row);
      piece.dataset.col = String(cell.col);
      square.appendChild(piece);
    }

    boardElement.appendChild(square);
  }
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
    if (turnIndicator) {
      turnIndicator.textContent = 'Unavailable';
    }
  }
}

async function selectPiece(pieceId, row, col) {
  selectedPieceId = pieceId;
  selectedSquareKey = squareKey(row, col);
  legalMoveSquareKeys = [];
  renderBoard();

  try {
    const data = await fetchJSON(`/api/legal-moves?pieceId=${encodeURIComponent(pieceId)}`, {
      headers: {
        Accept: 'application/json'
      }
    });

    if (selectedPieceId !== pieceId) {
      return;
    }

    legalMoveSquareKeys = Array.isArray(data.moves)
      ? data.moves.map((move) => squareKey(move.to.row, move.to.col))
      : [];
    renderBoard();
  } catch (error) {
    console.error('Failed to load legal moves', error);
    if (selectedPieceId === pieceId) {
      legalMoveSquareKeys = [];
      renderBoard();
    }
  }
}

if (boardElement) {
  boardElement.addEventListener('click', async (event) => {
    const square = event.target instanceof Element ? event.target.closest('.square') : null;
    if (square instanceof HTMLElement) {
      const pieceId = square.dataset.pieceId;
      const row = Number.parseInt(square.dataset.row ?? '', 10);
      const col = Number.parseInt(square.dataset.col ?? '', 10);

      if (pieceId && Number.isInteger(row) && Number.isInteger(col)) {
        await selectPiece(pieceId, row, col);
        return;
      }
    }

    clearSelection();
    renderBoard();
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
  if (event.key !== 'Escape' || !selectedPieceId) {
    return;
  }

  clearSelection();
  renderBoard();
});

void loadBoard();
