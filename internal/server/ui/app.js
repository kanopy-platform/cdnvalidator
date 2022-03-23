'use strict'

function apiPrefix() {
    return "/api/v1beta1/distributions"
}

async function getJson(url = "") {
    const response = await fetch(url, {
        method: "GET",
        headers: {
        "Accept": "application/json",
        }
    })
    .catch(error => {
        throw new Error(error)
    })

    const data = await response.json();

    if (!response.ok) {
        const message = `GET ${url} returned error: ${response.status} ${data.status}`;
        throw new Error(message)
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
        throw new Error(error)
    })

    const data = await response.json();

    if (!response.ok) {
        const message = `POST ${url} returned error: ${response.status} ${data.status}`;
        throw new Error(message)
    }
    return data;
}

async function populateDistributionDropdowns() {
    console.log("YUZHOU DEBUG in populateDistributionDropdowns")

    let createInvalidationDropdown = document.getElementById("create-invalidation-distribution")
    let getInvalidationDropdown = document.getElementById("get-invalidation-distribution")

    await getJson(apiPrefix())
    .then(data => {
        data.distributions.forEach(distribution => {
        let option = document.createElement("option")
        option.text = distribution
        option.value = distribution

        createInvalidationDropdown.appendChild(option.cloneNode(true))
        getInvalidationDropdown.appendChild(option.cloneNode(true))
        })
    })
    .catch(error => {
        console.log(error.message);
    })
}

async function createInvalidation() {
    console.log("YUZHOU DEBUG in createInvalidation")

    let distribution = document.getElementById("create-invalidation-distribution").value;
    let url = apiPrefix().concat("/", distribution, "/invalidations");

    let commaSeparatedPaths = document.getElementById("create-invalidation-paths").value;
    let pathsArr = commaSeparatedPaths.split(",").map(function(item) {
        return item.trim();
    });

    await postJson(url, { paths: pathsArr })
    .then(data => {
        console.log(data)
    })
    .catch(error => {
        console.log(error.message);
    });
}

async function getInvalidation() {
    console.log("YUZHOU DEBUG in getInvalidation")

    let distribution = document.getElementById("get-invalidation-distribution").value;
    let invalidationID = document.getElementById("get-invalidation-id").value;
    // remove surrounding quotes if any
    invalidationID = invalidationID.replace(/^"(.*)"$/, "$1");
    let url = apiPrefix().concat("/", distribution, "/invalidations/", invalidationID);

    console.log("YUZHOU DEBUG invalidationID: " + invalidationID)

    await getJson(url)
    .then(data => {
        console.log(data)
    })
    .catch(error => {
        console.log(error.message);
    });
}
