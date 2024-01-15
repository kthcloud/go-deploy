const mockData = [
    {
        "name": "confirmer",
        "status": "running",
    },
    {
        "name": "logger",
        "status": "stopped",
    },
    {
        "name": "repairer",
        "status": "running",
    },
    {
        "name": "pinger",
        "status": "stopped",
    }
];

function sentenceCase(str) {
    const result = str.replace(/([A-Z])/g, " $1");
    return result.charAt(0).toUpperCase() + result.slice(1);
}

function createBox(name, status) {
    const boxDiv = document.createElement('div');
    boxDiv.classList.add('box');

    const ledDiv = document.createElement('div');
    ledDiv.classList.add('led', status); // Add the status as a class for color
    ledDiv.style.animation = 'blink-once 1s'; // Add blink animation

    const nameDiv = document.createElement('div');
    nameDiv.classList.add('name');
    nameDiv.textContent = name;

    const statusDiv = document.createElement('div');
    statusDiv.classList.add('status');
    statusDiv.textContent = status;

    boxDiv.appendChild(ledDiv);
    boxDiv.appendChild(nameDiv);
    boxDiv.appendChild(statusDiv);

    return boxDiv;
}

function loadBoxes() {
    const container = document.getElementById('boxes-container');

    fetch(apiURL).then(response => {
        return response.json();
    }).then(data => {
        container.innerHTML = ''; // Clear the container
        data.forEach(data => {
            const box = createBox(sentenceCase(data.name), sentenceCase(data.status));
            container.appendChild(box);
        });
    });
}

// Initial load of boxes
loadBoxes();

// Reload boxes every 5 seconds
function reloadBoxes() {
    loadBoxes();
}

setInterval(reloadBoxes, 2000);