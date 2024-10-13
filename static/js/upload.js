let sessionToken = "";
let showSessionToken = false

const showSessionTokenBtn = document.getElementById("show-session-token-btn")
const changeSessionTokenBtn = document.getElementById("change-session-token-btn")
const sessionTokenDisplay = document.getElementById("session-token-display")

function displaySessionToken() {
  if (!sessionToken) {
    sessionTokenDisplay.innerText = "<No Token>"
  } else if (showSessionToken) {
    sessionTokenDisplay.innerText = sessionToken
  } else {
    sessionTokenDisplay.innerText = "*".repeat(sessionToken.length)
  }
}

showSessionTokenBtn.addEventListener("click", () => {
  if (showSessionToken) {
    showSessionTokenBtn.innerText = "Show"
  } else {
    showSessionTokenBtn.innerText = "Hide"
  }
  showSessionToken = !showSessionToken

  displaySessionToken()
})

changeSessionTokenBtn.addEventListener("click", () => {
  input = prompt("Session Token:")
  if (input !== sessionToken) {
    sessionToken = input
    displaySessionToken()
  }
})


const fileForm = document.getElementById("file-form");
const messageForm = document.getElementById("message-form");
const uploadResult = document.getElementById("upload-result");

fileForm.addEventListener("submit", e => {
  e.preventDefault();

  const formData = new FormData(fileForm);

  while (!sessionToken) {
    const input = prompt("Session Token:")
    if (input === null) {
      return
    }
    sessionToken = input
    displaySessionToken()
  }


  fetch("/api/upload", {
    method: "POST",
    headers: {
      "Authorization": `Bearer ${sessionToken}`
    },
    body: formData
  })
    .then(async res => {
      if (res.status >= 400) {
        throw new Error(await res.text())
      }
      uploadResult.innerText = await res.text();
      setTimeout(() => {
        uploadResult.innerText = ""
      }, 2000);
      fileForm.reset();
    })
    .catch(err => {
      uploadResult.innerText = `Error when uploading files: ${err}`;
      console.error("Error when uploading files:", err)
    });

});

messageForm.addEventListener("submit", e => {
  e.preventDefault();

  const formData = new FormData(messageForm);
  const body = formData.get("message");
  const timeSent = new Date();

  while (!sessionToken) {
    const input = prompt("Session Token:")
    if (input === null) {
      return
    }
    sessionToken = input
    displaySessionToken()
  }

  fetch("/api/message", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Authorization": `Bearer ${sessionToken}`
    },
    body: JSON.stringify({ body, timeSent }),
  })
    .then(async res => {
      if (res.status >= 400) {
        throw new Error(await res.text())
      }
      uploadResult.innerText = await res.text();
      setTimeout(() => {
        uploadResult.innerText = ""
      }, 2000);
      messageForm.reset();
    })
    .catch(err => {
      uploadResult.innerText = `Error when sending message: ${err}`;
      console.error("Error when sending message:", err);
    });

});

sessionToken = prompt("Session Token:")
displaySessionToken()
