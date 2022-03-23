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
    .catch(error => {
        throw new Error(error);
    })

    const data = await response.json();

    if (!response.ok) {
        const message = `GET ${url} returned error: ${response.status} ${data.status}`;
        throw new Error(message);
    }
    return data;
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
    .catch(error => {
        throw new Error(error);
    })

    const data = await response.json();

    if (!response.ok) {
        const message = `POST ${url} returned error: ${response.status} ${data.status}`;
        throw new Error(message);
    }
    return data;
}

function appendOutput(header = "", details = "", data = {}) {
    // Add item to side nav bar
    let navBar = document.getElementById("nav-bar");

    let linkNode = document.createElement("a");
    linkNode.setAttribute("class", "list-group-item list-group-item-action");
    linkNode.href = "#" + itemPrefix + itemCount;
    linkNode.innerHTML = header;

    navBar.appendChild(linkNode);

    // Add to output box
    let outputBox = document.getElementById("output-box");

    let headerNode = document.createElement("h4");
    headerNode.id = itemPrefix.concat(itemCount);
    headerNode.innerHTML = header;

    let paragraphNode = document.createElement("p")
    paragraphNode.innerHTML = details;

    let preNode = document.createElement("pre");
    preNode.textContent = JSON.stringify(data, undefined, 2);

    outputBox.appendChild(headerNode);
    outputBox.appendChild(paragraphNode);
    outputBox.appendChild(preNode);

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
    // show loading spinner
    document.getElementById("loading").style.visibility="visible";

    // construct the POST request
    let distribution = document.getElementById("create-invalidation-distribution").value;
    let url = apiPrefix.concat("/", distribution, "/invalidations");

    let commaSeparatedPaths = document.getElementById("create-invalidation-paths").value;
    let pathsArr = commaSeparatedPaths.split(",").map(function(item) {
        return item.trim().replace(/^"(.*)"$/, "$1");   // remove surrounding whitespace and quotes if any
    });

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

    // hide loading spinner
    document.getElementById("loading").style.visibility="hidden";
}



async function getInvalidation() {
    // show loading spinner
    document.getElementById("loading").style.visibility="visible";

    // construct the GET request
    let distribution = document.getElementById("get-invalidation-distribution").value;
    let invalidationID = document.getElementById("get-invalidation-id").value;
    invalidationID = invalidationID.replace(/^"(.*)"$/, "$1");  // remove surrounding quotes if any
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

    // hide loading spinner
    document.getElementById("loading").style.visibility="hidden";
}
