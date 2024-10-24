const sharedDirContainer = document.getElementById("shared-dirs-container");
const refreshBtn = document.getElementById("shared-dirs-refresh-btn");

refreshBtn.addEventListener("click", async () => {
  while (!sessionToken) {
    const input = prompt("Session Token:");
    if (input === null) {
      return;
    }
    sessionToken = input;
    displaySessionToken();
  }
  await refreshSharedDirs();
});

// Auto-fetch on load
refreshSharedDirs();

async function refreshSharedDirs() {
  try {
    const rootDirs = await fetchRootSharedDirs();
    const rootDirElements = rootDirs.map(createSharedFileElement);
    sharedDirContainer.replaceChildren(...rootDirElements);
  } catch (e) {
    console.error(e);
    sharedDirContainer.replaceChildren("Failed to load shared directories");
  }
}

/**
 * @typedef {Object} SharedFile
 * @property {number} fType
 * @property {string} name
 * @property {string} path
 * @property {number} size
 * @property {string} rootDirHash
 * @property {SharedFile[]} children
 */

/**
 * Fetches the root shared directories
 * @returns list of root shared directories
 */
async function fetchRootSharedDirs() {
  return fetchDirectoryContent("", "");
}

/**
 * @param {string} path relative path from root dir
 * @param {string} rootDirHash  hash of the root directory
 * @returns {Promise<SharedFile>}
 */
async function fetchDirectoryContent(path, rootDirHash) {
  const params = new URLSearchParams({
    "root-dir-hash": rootDirHash,
  });

  if (path) {
    params.append("path", path);
  }

  return await fetch(`/api/shared-dir?${params.toString()}`, {
    headers: {
      Authorization: `Bearer ${sessionToken}`,
    },
  }).then((res) => res.json());
}

const FILE = 0;
const DIRECTORY = 1;

/**
 * @param {SharedFile} file
 * @returns {HTMLElement}
 * @throws Throws error when `file.fType` is unknown
 */
function createSharedFileElement(file) {
  console.debug("createSharedFileElement for", file);

  let element = null;

  if (file.fType === FILE) {
    const div = document.createElement("div");
    const span = document.createElement("span");
    span.classList.add("shared-dir-shared-file");

    span.innerText = file.name;
    span.addEventListener("click", () => handleFileDownload(file));

    div.appendChild(span);

    element = div;
  } else if (file.fType === DIRECTORY) {
    const details = document.createElement("details");
    const summary = document.createElement("summary");
    summary.innerText = file.name;

    const downloadBtn = document.createElement("button");
    downloadBtn.innerText = "⬇";
    downloadBtn.classList.add("shared-dir-download-btn");
    downloadBtn.classList.add("hidden");
    downloadBtn.addEventListener("click", () => handleFileDownload(file));

    summary.appendChild(downloadBtn);

    // Hide downloadBtn until summary is hovered
    summary.addEventListener("mouseenter", () => {
      downloadBtn.classList.remove("hidden");
    });
    summary.addEventListener("mouseleave", () => {
      downloadBtn.classList.add("hidden");
    });

    const childrenContainer = document.createElement("div");
    childrenContainer.classList.add("shared-dir-children-container");
    if (file.children) {
      for (const child of file.children) {
        childrenContainer.appendChild(createSharedFileElement(child));
      }
    }

    details.appendChild(summary);
    details.appendChild(childrenContainer);

    summary.addEventListener("click", async (e) => {
      e.stopPropagation();

      // This either means that the directory is empty or that it wasn't fetched.
      // Implementing a check to avoid extra network calls when the directory is
      // actually empty can be an improvement.
      //
      // Fetches the content of the directory and adds them to the DOM.
      if (childrenContainer.children.length === 0) {
        let sharedDir;
        try {
          sharedDir = await fetchDirectoryContent(file.path, file.rootDirHash);
        } catch (e) {
          console.error(`Failed to fetch directory content of ${file.path}`, e);
          return;
        }

        // BUG: sharedDir is `undefined` when sharing uploaded directory
        for (const child of sharedDir.children) {
          childrenContainer.appendChild(createSharedFileElement(child));
        }
      }
    });

    element = details;
  } else {
    throw new Error(`Unknown file type: ${file.fType}`);
  }

  return element;
}

/**
 * @param {SharedFile} file SharedFile to download
 */
async function handleFileDownload(file) {
  console.log(`Trying to download ${file.name}`);

  let response;
  try {
    response = await fetchFileBinaryContent(file);
    console.log("Fetched binary content");
  } catch (e) {
    console.error(`Failed to download file: ${file.name}`, e);
    return;
  }

  if (!response.ok) {
    console.error(
      `Failed to download file: ${file.name}`,
      await response.text()
    );
    return;
  }

  const fileSize = response.headers.get("content-length")
  console.info(`File size: ${fileSize}bytes`)

  const blob = await response.blob();

  console.log("Creating anchor element");
  const a = document.createElement("a");
  a.download = file.name;
  const url = window.URL.createObjectURL(blob);
  a.href = url;

  document.body.appendChild(a);
  console.log("Starting download");
  a.click();
  document.body.removeChild(a);

  window.URL.revokeObjectURL(url);

  // TODO: Add progress bar
  // const reader = response.body.getReader()
  // while (true) {
  //   const { value, done } = await reader.read()
  //   if (done) {
  //     break;
  //   }
  //
  // }
}

/**
 * @param {SharedFile} file
 * @returns The `Response` object from server containing the file download or error
 */
async function fetchFileBinaryContent(file) {
  const params = new URLSearchParams({
    path: file.path,
    "root-dir-hash": file.rootDirHash,
  });

  return await fetch(`/api/download?${params.toString()}`, {
    headers: {
      Authorization: `Bearer ${sessionToken}`,
    },
  });
}
