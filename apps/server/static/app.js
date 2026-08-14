const newGameButton = document.getElementById('new-game-button');
const turnIndicator = document.getElementById('turn-indicator');

if (newGameButton && turnIndicator) {
  newGameButton.addEventListener('click', async () => {
    newGameButton.disabled = true;

    try {
      const response = await fetch('/games/new', {
        method: 'POST',
        headers: {
          'Accept': 'application/json'
        }
      });

      if (!response.ok) {
        throw new Error('request failed');
      }

      const data = await response.json();
      if (typeof data.turn === 'string') {
        turnIndicator.textContent = data.turn.charAt(0).toUpperCase() + data.turn.slice(1);
      }
    } catch (error) {
      console.error('Failed to start new game', error);
    } finally {
      newGameButton.disabled = false;
    }
  });
}
