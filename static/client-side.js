const editor = document.getElementById("editor");
let timeout;

fetch("/api/text")
  .then((res) => res.text())
  .then((text) => (editor.value = text));

editor.addEventListener("input", () => {
  clearTimeout(timeout);
  timeout = setTimeout(() => {
    fetch("/api/text", {
      method: "POST",
      body: editor.value,
    });
  }, 500);
});
