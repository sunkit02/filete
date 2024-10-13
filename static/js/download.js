const sharedDirContainer = document.getElementById("shared-dirs-container");
const refreshBtn = document.getElementById("shared-dirs-refresh-btn");

refreshBtn.addEventListener("click", async () => {
  while (!sessionToken) {
    const input = prompt("Session Token:")
    if (input === null) {
      return
    }
    sessionToken = input
    displaySessionToken()
  }
  await refreshSharedDirs();
});

// Auto-fetch on load
refreshSharedDirs()

async function refreshSharedDirs() {
  try {
    const rootDirs = await fetchRootSharedDirs();
    const rootDirElements = rootDirs.map(createSharedFileElement);
    sharedDirContainer.replaceChildren(...rootDirElements);
  } catch (e) {
    console.error(e);
    sharedDirContainer.replaceChildren("Failed to load shared directories")
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
 * @returns {Promise<SharedFile[]>}
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
  console.debug("createSharedFileElement for", file)

  let element = null;

  if (file.fType === FILE) {
    const div = document.createElement("div");
    const a = document.createElement("a");

    a.innerText = file.name;

    div.appendChild(a);

    element = div;
  } else if (file.fType === DIRECTORY) {
    const details = document.createElement("details");
    const summary = document.createElement("summary");
    summary.innerText = file.name;

    const childrenContainer = document.createElement("div");
    childrenContainer.classList.add("shared-dir-children-container");
    for (const child of file.children) {
      childrenContainer.appendChild(createSharedFileElement(child));
    }

    details.appendChild(summary);
    details.appendChild(childrenContainer);

    element = details;
  } else {
    throw new Error(`Unknown file type: ${file.fType}`);
  }

  return element;
}
