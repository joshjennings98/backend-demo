const iframe = document.getElementById('command-display');
const outputDiv = document.getElementById('text-output');
const pageSelect = document.getElementById('page-select');
const commandControlWrapper = document.getElementById('command-control-wrapper');
const refreshCheckbox = document.getElementById('auto-refresh-command');
const secondsDropdown = document.getElementById('seconds');
const currentCommand = document.getElementById('current-command');

let pages = [];
let textLines = [];
let lineIndex = 0;
let displayedLines = [];
let index = 0;
let navigationDirection = 'forward';
var myInterval;

async function fetchPages() {
    try {
        const response = await fetch('/pages');
        const data = await response.json();
        pages = data.pages;

        // Generate select options
        pages.forEach((_, i) => {
            const option = document.createElement('option');
            option.value = i;
            option.textContent = `Slide ${i + 1}/${pages.length}`;
            pageSelect.appendChild(option);
        });

        loadPage();
    } catch (error) {
        console.error("Error fetching pages: ", error);
    }
}

async function loadPage() {
    try {
        // 'about:blank' is always blank so we can use it to clear the iframe (prevents ghosting during command switch)
        iframe.src = "about:blank"
        refreshCheckbox.checked = false;
        clearInterval(myInterval); 

        // Fetch and display current command
        const commandResponse = await fetch(`/command/${index}`);
        const commandData = await commandResponse.text();
        currentCommand.innerHTML = commandData;

        if (pages[index].type === commandTag) {
            outputDiv.style.display = 'none';
            iframe.style.display = 'block';
            commandControlWrapper.style.display = 'flex';
            currentCommand.style.display = 'block';
            textLines = [];
            displayedLines = [];
            lineIndex = 0;
        } else if (pages[index].type === codeTag || pages[index].type === imageTag) {
            iframe.style.display = 'none';
            currentCommand.style.display = 'none';
            commandControlWrapper.style.display = 'none';
            const response = await fetch(`/pages/${index}`);
            const data = await response.text();
            outputDiv.innerHTML = data;
            if (pages[index].type === codeTag) {
                outputDiv.style.height = ``;
                hljs.highlightAll();
            }
            outputDiv.style.display = 'block';
        } else {
            iframe.style.display = 'none';
            currentCommand.style.display = 'none';
            commandControlWrapper.style.display = 'none';
            const response = await fetch(`/pages/${index}`);
            const data = await response.json();
            textLines = data.content;
            if (navigationDirection === 'backward') {
                // If we're navigating backward, display all lines
                lineIndex = textLines.length - 1;
                displayedLines = [...textLines];
            } else {
                // If we're navigating forward, display the first line
                lineIndex = 0;
                displayedLines = [textLines[lineIndex]];
            }
            outputDiv.innerHTML = displayedLines.join('<br><br><br>');
            outputDiv.style.height = `${3 * textLines.length}em`;
            outputDiv.style.display = 'block';
        }

        pageSelect.value = index;
    } catch (error) {
        console.error("Error loading page: ", error);
    }
}

function pageBackwards() {
    if (pages[index].type === textTag && displayedLines.length > 1) {
        displayedLines.pop();
        lineIndex--;
        outputDiv.innerHTML = displayedLines.join('<br><br><br>');
    } else if (index > 0) {
        index--;
        navigationDirection = 'backward';
        loadPage().catch(console.error);
    }
}

function pageForwards() {
    if (pages[index].type === textTag && lineIndex < textLines.length - 1) {
        lineIndex++;
        displayedLines.push(textLines[lineIndex]);
        outputDiv.innerHTML = displayedLines.join('<br><br><br>');
    } else if (index < pages.length - 1) {
        index++;
        navigationDirection = 'forward';
        loadPage().catch(console.error);
    }
}

function runCommand() {
    iframe.src = `/pages/${index}`;
}

function refreshCommand() {
    clearInterval(myInterval); 
    if (refreshCheckbox.checked) {
        myInterval = setInterval(runCommand, `${secondsDropdown.value}000`);
    }
}

secondsDropdown.addEventListener('change', () => {
    refreshCheckbox.checked = true;
    refreshCommand();
});

refreshCheckbox.addEventListener("change", function() {
    refreshCommand();
});

pageSelect.addEventListener('change', () => {
    index = Number(pageSelect.value);
    loadPage().catch(console.error);
});

document.body.addEventListener('mousedown', (event) => {
    const controlsDiv = document.getElementById('controls');
    let target = event.target;

    while (target != null) {
        if (target === controlsDiv || target === commandControlWrapper || (target.tagName != null && target.tagName.toLowerCase() === "a")) {
            return;
        }
        target = target.parentNode;
    }

    if (event.button === 0) { // Left click
        pageForwards();
    } else if (event.button === 2) { // Right click
        pageBackwards();
    }
});

document.body.addEventListener('keydown', (event) => {
    if (event.key === "ArrowRight") {
        pageForwards();
    } else if (event.key === "ArrowLeft") {
        pageBackwards();
    } else if (event.key === " " && pages[index].type === commandTag) {
        runCommand();
    } else if (event.key === "r" && pages[index].type === commandTag) {
        refreshCheckbox.checked = !refreshCheckbox.checked;
        refreshCommand();
    }
});

document.body.addEventListener('contextmenu', (event) => {
    event.preventDefault(); // Prevent the context menu from showing on right-click
});

fetchPages().catch(console.error); // Fetch pages on load
