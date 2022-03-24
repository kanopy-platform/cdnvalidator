"use strict";

let itemCount = 0;
const apiPrefix = "/api/v1beta1/distributions";
const itemPrefix = "item-";

async function getJson(url = "") {
    const response = await fetch(url, {
        method: "GET",
        headers: {
        "Accept": "application/json",
        }
    })

    if (response.ok) {
        let data = await response.json();
        return data;
    }

    const responseText = await response.text();
    let message;
    try {
        message = JSON.parse(responseText).status;
    } catch {
        message = `GET ${url} returned error: ${response.status} ${response.statusText}: ${responseText}`;
    }
    throw new Error(message);
}

async function postJson(url = "", postData = {}) {
    const response = await fetch(url, {
        method: "POST",
        headers: {
        "Accept": "application/json",
        "Content-Type": "application/json",
        },
        body: JSON.stringify(postData),
    })

    if (response.ok) {
        let data = await response.json();
        return data;
    }

    const responseText = await response.text();
    let message;
    try {
        message = JSON.parse(responseText).status;
    } catch {
        message = `POST ${url} returned error: ${response.status} ${response.statusText}: ${responseError}`;
    }
    throw new Error(message);
}

function getTime() {
    let d = new Date();
    return d.toLocaleTimeString();
}

function appendOutput(header = "", details = "", data = undefined) {
    // Add timestamp to header
    header = header + " : " + getTime();

    // Add item to top of side nav bar
    let navBar = document.getElementById("nav-bar");
    // Remove current active items
    let activeElems = navBar.querySelector(".active");
    if (activeElems !== null) {
        activeElems.classList.remove("active");
    }

    let linkNode = document.createElement("a");
    linkNode.setAttribute("class", "list-group-item list-group-item-action active");
    linkNode.href = "#" + itemPrefix + itemCount;
    linkNode.innerHTML = header;

    navBar.insertBefore(linkNode, navBar.firstChild);
    // Reset scroll to top
    navBar.scrollTop = 0;

    // Add to top of output box
    let outputBox = document.getElementById("output-box");

    let headerNode = document.createElement("h4");
    headerNode.id = itemPrefix.concat(itemCount);
    headerNode.innerHTML = header;

    let paragraphNode = document.createElement("p")
    paragraphNode.innerHTML = details;

    let preNode = document.createElement("pre");
    if (data !== undefined) {
        preNode.textContent = JSON.stringify(data, undefined, 2);
    }

    // Must insert in reverse order of display
    outputBox.insertBefore(preNode, outputBox.firstChild);
    outputBox.insertBefore(paragraphNode, outputBox.firstChild);
    outputBox.insertBefore(headerNode, outputBox.firstChild);
    // Reset scroll to top
    outputBox.scrollTop = 0;

    itemCount++;

    $('[data-spy="scroll"]').each(function () {
        $(this).scrollspy('refresh');
    })
}

async function populateDistributionDropdowns() {
    let createInvalidationDropdown = document.getElementById("create-invalidation-distribution");
    let getInvalidationDropdown = document.getElementById("get-invalidation-distribution");

    await getJson(apiPrefix)
    .then(data => {
        data.distributions.forEach(distribution => {
            let option = document.createElement("option");
            option.text = distribution;
            option.value = distribution;

            createInvalidationDropdown.appendChild(option.cloneNode(true));
            getInvalidationDropdown.appendChild(option.cloneNode(true));
        })
    })
    .catch(error => {
        appendOutput("Error Populating Distributions", "", error.message);
    })

    document.getElementById("loading").style.visibility="hidden";
}

async function createInvalidation() {
    // construct the POST request
    let distribution = document.getElementById("create-invalidation-distribution").value;
    let url = apiPrefix.concat("/", distribution, "/invalidations");

    let commaSeparatedPaths = document.getElementById("create-invalidation-paths").value;
    let pathsArr = commaSeparatedPaths.split(",").map(function(item) {
        return item.trim().replace(/^"(.*)"$/, "$1");   // remove surrounding whitespace and quotes if any
    });

    if (pathsArr.length == 1 && pathsArr[0] == "") {
        appendOutput("Create Error", "Paths is empty");
        return;
    }

    // disable submit button, show loading spinner
    document.getElementById("create-invalidation-button").disabled = true;
    document.getElementById("loading").style.visibility="visible";

    let details = "<b>Distribution:</b> " + distribution + "<br />";
    details = details + "<b>Paths:</b> " + pathsArr.join(",") + "<br />";

    // send POST request
    await postJson(url, { paths: pathsArr })
    .then(data => {
        appendOutput("Create", details, data);
    })
    .catch(error => {
        appendOutput("Create Error", details, error.message);
    });

    // enable submit button, hide loading spinner
    document.getElementById("create-invalidation-paths").value = "";
    document.getElementById("create-invalidation-button").disabled = false;
    document.getElementById("loading").style.visibility="hidden";
}



async function getInvalidation() {
    // construct the GET request
    let distribution = document.getElementById("get-invalidation-distribution").value;
    let invalidationID = document.getElementById("get-invalidation-id").value;
    invalidationID = invalidationID.replace(/^"(.*)"$/, "$1");  // remove surrounding quotes if any
    if (invalidationID == "") {
        appendOutput("Get Error", "Invalidation ID is empty");
        return;
    }

    // disable submit button, show loading spinner
    document.getElementById("get-invalidation-button").disabled = true;
    document.getElementById("loading").style.visibility="visible";

    let url = apiPrefix.concat("/", distribution, "/invalidations/", invalidationID);
    
    let details = "<b>Distribution:</b> " + distribution + "<br />";
    details = details + "<b>Invalidation ID:</b> " + invalidationID + "<br />";

    // send GET request
    await getJson(url)
    .then(data => {
        appendOutput("Get", details, data);
    })
    .catch(error => {
        appendOutput("Get Error", details, error.message);
    });

    // enable submit button, hide loading spinner
    document.getElementById("get-invalidation-id").value = "";
    document.getElementById("get-invalidation-button").disabled = false;
    document.getElementById("loading").style.visibility="hidden";
}
