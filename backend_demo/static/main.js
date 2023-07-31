const iframe = document.getElementById('command-display');
const outputDiv = document.getElementById('text-output');
const commandSelect = document.getElementById('command-select');
const currentCommand = document.getElementById('current-command');

let commands = [];
let textLines = [];
let lineIndex = 0;
let displayedLines = [];
let index = 0;
let navigationDirection = 'forward';

async function fetchCommands() {
    try {
        const response = await fetch('/pages');
        const data = await response.json();
        commands = data.commands;

        // Generate select options
        commands.forEach((command, i) => {
            const option = document.createElement('option');
            option.value = i;
            option.textContent = `Slide ${i + 1}/${commands.length}`;
            commandSelect.appendChild(option);
        });

        await loadCommand();
    } catch (error) {
        console.error('Error:', error);
    }
}

async function loadCommand() {
    try {
        outputDiv.style.display = 'none';
        iframe.style.display = 'none';
        currentCommand.style.display = 'none';

        if (commands[index].type === commandTag) {
            iframe.src = `/pages/${index}`;
            iframe.style.display = 'block';
            currentCommand.style.display = 'block';
            textLines = [];
            displayedLines = [];
            lineIndex = 0;
        } else if (commands[index].type === codeTag || commands[index].type === imageTag) {
            const response = await fetch(`/pages/${index}`);
            const data = await response.text();
            outputDiv.innerHTML = data;
            if (commands[index].type === codeTag) {
                outputDiv.style.height = ``;
                hljs.highlightAll();
            }
            outputDiv.style.display = 'block';
        } else {
            const response = await fetch(`/pages/${index}`);
            const data = await response.json();
            textLines = data.text_lines;
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

        commandSelect.value = index;
        currentCommand.textContent = "";

        // Fetch and display current command
        const commandResponse = await fetch(`/command/${index}`);
        const commandData = await commandResponse.text();
        currentCommand.textContent = commandData;
    } catch (error) {
        console.error('Error:', error);
    }
}

commandSelect.addEventListener('change', () => {
    index = Number(commandSelect.value);
    loadCommand().catch(console.error);
});

document.body.addEventListener('mousedown', (event) => {
    const controlsDiv = document.getElementById('controls');
    let target = event.target;

    while (target != null) {
        if (target === controlsDiv || (target.tagName != null && target.tagName.toLowerCase() === "a")) {
            return;
        }
        target = target.parentNode;
    }

    if (event.button === 2) { // Right-click
        handleRightClick();
    } else if (event.button === 0) { // Left-click
        handleLeftClick();
    }
});

// Prevent the context menu from showing on right-click
document.body.addEventListener('contextmenu', (event) => {
    event.preventDefault();
});

// Fetch commands on load
fetchCommands();

// Right click to go backwards
function handleRightClick() {
    try {
        if (commands[index].type === textTag && displayedLines.length > 1) {
            displayedLines.pop();
            lineIndex--;
            outputDiv.innerHTML = displayedLines.join('<br><br><br>');
        } else if (index > 0) {
            index--;
            navigationDirection = 'backward';
            loadCommand().catch(console.error);
        }
    } catch (error) {
        console.error('Error:', error);
    }
}

// Left click to go forwards
function handleLeftClick() {
    try {
        if (commands[index].type === textTag && lineIndex < textLines.length - 1) {
            lineIndex++;
            displayedLines.push(textLines[lineIndex]);
            outputDiv.innerHTML = displayedLines.join('<br><br><br>');
        } else if (index < commands.length - 1) {
            index++;
            navigationDirection = 'forward';
            loadCommand().catch(console.error);
        }
    } catch (error) {
        console.error('Error:', error);
    }
}
