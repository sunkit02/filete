let sessionKey = "";
let showSessionKey = false

const showSessionKeyBtn = document.getElementById("show-session-key-btn")
const changeSessionKeyBtn = document.getElementById("change-session-key-btn")
const sessionKeyDisplay = document.getElementById("session-key-display")
const authenticateBtn = document.getElementById("authenticate-btn")
const endSessionBtn = document.getElementById("end-session-btn")

authenticateBtn.addEventListener("click", () => {
  while (!sessionKey) {
    const input = prompt("Session Key:")
    if (input === null) {
      return
    }
    sessionKey = input
    displaySessionKey()
  }

  fetch("/auth/authenticate", {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify({ sessionKey })
  })
    .then(async (res) => {
      if (!res.ok) {
        const msg = `Failed to authenticate: ${await res.text()}`
        alert(msg)
        console.error(msg)
      }
    })
    .catch(err => console.error(err))
});
endSessionBtn.addEventListener("click", () => {
  fetch("/auth/invalidate-token", {
    method: "POST",
  })
    .then(async (res) => {
      if (!res.ok) {
        throw new Error(await res.text())
      }
    })
    .catch(err => console.error(err))
})

function displaySessionKey() {
  if (!sessionKey) {
    sessionKeyDisplay.innerText = "<No Key>"
  } else if (showSessionKey) {
    sessionKeyDisplay.innerText = sessionKey
  } else {
    sessionKeyDisplay.innerText = "*".repeat(sessionKey.length)
  }
}

showSessionKeyBtn.addEventListener("click", () => {
  if (showSessionKey) {
    showSessionKeyBtn.innerText = "Show"
  } else {
    showSessionKeyBtn.innerText = "Hide"
  }
  showSessionKey = !showSessionKey

  displaySessionKey()
})

changeSessionKeyBtn.addEventListener("click", () => {
  input = prompt("Session Key:")
  if (input !== sessionKey) {
    sessionKey = input
    displaySessionKey()
  }
})


const fileForm = document.getElementById("file-form");
const messageForm = document.getElementById("message-form");
const uploadResult = document.getElementById("upload-result");

fileForm.addEventListener("submit", e => {
  e.preventDefault();

  const formData = new FormData(fileForm);

  while (!sessionKey) {
    const input = prompt("Session Key:")
    if (input === null) {
      return
    }
    sessionKey = input
    displaySessionKey()
  }


  fetch("/api/upload", {
    method: "POST",
    headers: {
      "Authorization": `Bearer ${sessionKey}`
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

  while (!sessionKey) {
    const input = prompt("Session Key:")
    if (input === null) {
      return
    }
    sessionKey = input
    displaySessionKey()
  }

  fetch("/api/message", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Authorization": `Bearer ${sessionKey}`
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

sessionKey = prompt("Session Key:")
displaySessionKey()
