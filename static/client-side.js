const editor = document.getElementById("editor");

const ws = new WebSocket(`ws://${location.host}/ws`);

ws.addEventListener("message", (event) => {
  if (editor.value !== event.data) {
    editor.value = event.data;
  }
});

editor.addEventListener("input", () => {
  if (ws.readyState === WebSocket.OPEN) {
    ws.send(editor.value);
  }
});
