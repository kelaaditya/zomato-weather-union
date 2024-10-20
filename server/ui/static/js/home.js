//
// process data
//
// 'data' variable is loaded with data fetched from the server
const dataProcessed = data.map((element) => {
    // convert timestamp from server to locale string
    const timeOfCalculation = new Date(element.time_stamp_calculation)
    const timeString = timeOfCalculation.toLocaleString("en-IN", { "hour12": false })

    // append new time string to data element
    element.time_string = timeString

    return element
});

//
// information
//
let information = document.getElementById("information")
information.textContent = `
    Showing ${dataProcessed.length} measurements below.\n
    The time of calculation of this run was approximately ${dataProcessed[0].time_string} IST (DD-MM-YYYY).
`

//
// colour bar
//
const colourBar = document.getElementById("colour-bar");

// fill the colour bar with gradients
for (let i = 20; i <= 40; i++) {
    const listItem = document.createElement('li');
    listItem.innerHTML = `
        <div class="colour-box"></div>
        <div class="colour-box-item-value">${
            (i == 20) ? "≤" + i + "°C" : (i == 40) ? "≥" + i + "°C": i + "°C"
        }</div>
    `;
    colourBar.appendChild(listItem);
}

// set colour here as inline styles are not supported by the CSP
// set by the server
// <div class="colour-box" style="background-color: ${color};"></div>
for (let i = 0; i <= 20; i++) {
    const colour = getGradientColor(i + 20);
    colourBar.children[i].children[0].style.backgroundColor = colour;
}

//
// map
//
// map centered on India
let map = L.map("map").setView([20, 78], 5);

// add a title layer with the Open Street Map information
L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
}).addTo(map);

// run the circle display function for the map
displayDataAsCircles(dataProcessed)

// function to set up measurements as circles via Leaflet
function displayDataAsCircles(dataArray) {
    for (let i = 0; i < dataArray.length; i++) {
        const {
            locality_id,
            locality_name,
            longitude,
            latitude,
            temperature_wet_bulb,
            time_string
        } = dataArray[i];

        // colour based in temperature
        const color = getGradientColor(temperature_wet_bulb);

        const circleMarker = L.circleMarker([latitude, longitude], {
            color: color,
            fillColor: color,
            fillOpacity: 0.6,
            radius: 9, // constant radius
        }).addTo(map);

        // add popup with the information
        circleMarker.on("click", function () {
            // // if there's an open popup, close it before opening a new one
            // if (openPopup) {
            //     openPopup.closePopup();
            // }

            // create the popup content
            const popupContent = `
            <div>
                <strong>Locality Name:</strong> ${locality_name}<br/>
                <strong>Locality ID:</strong> ${locality_id}<br/>
                <strong>Wet Bulb Temperature:</strong> ${temperature_wet_bulb}°C<br/>
                <strong>Time:</strong> ${time_string}
            </div>`;

            // bind popup to the marker and open it
            circleMarker.bindPopup(popupContent).openPopup();
            // openPopup = circleMarker;
        });
    }
}

// function to get colour values (rgb) for temperature values
// 20 is green, 30 is yellow and 40 is red.
// with a gradient in the middle
function getGradientColor(value) {
    const normalizedValue = (value - 20) / 20; // Normalize to 0-1 range
    if (normalizedValue <= 0.5) {
        // Green to Yellow
        const green = 255;
        const red = Math.round(normalizedValue * 2 * 255);
        return `rgb(${red}, ${green}, 0)`;
    } else {
        // Yellow to Red
        const red = 255;
        const green = Math.round((1 - (normalizedValue - 0.5) * 2) * 255);
        return `rgb(${red}, ${green}, 0)`;
    }
}
function getGradientColor(value) {
    // normalize to 0-1 range
    const normalizedValue = (value - 20) / 20;
    let r, g, b;
    if (normalizedValue < 0.33) {
        // blue to green
        r = 0;
        g = Math.round((normalizedValue / 0.33) * 255);
        b = Math.round((1 - normalizedValue / 0.33) * 255);
    } else if (normalizedValue < 0.67) {
        // green to yellow
        r = Math.round(((normalizedValue - 0.33) / 0.34) * 255);
        g = 255;
        b = 0;
    } else {
        // yellow to red
        r = 255;
        g = Math.round((1 - (normalizedValue - 0.67) / 0.33) * 255);
        b = 0;
    }

    return `rgb(${r}, ${g}, ${b})`;
}

// function getGradientColor(value) {
//     if (value <= 20) return "rgb(0, 255, 0)";
//     else if (value <= 21) return "rgb(10, 255, 0)";
//     else if (value <= 22) return "rgb(20, 255, 0)";
//     else if (value <= 23) return "rgb(30, 255, 0)";
//     else if (value <= 24) return "rgb(40, 255, 0)";
//     else if (value <= 25) return "rgb(51, 255, 0)";
//     else if (value <= 26) return "rgb(61, 255, 0)";
//     else if (value <= 27) return "rgb(71, 255, 0)";
//     else if (value <= 28) return "rgb(81, 255, 0)";
//     else if (value <= 29) return "rgb(91, 255, 0)";
//     else if (value <= 30) return "rgb(102, 255, 0)";
//     else if (value <= 31) return "rgb(112, 255, 0)";
//     else if (value <= 32) return "rgb(122, 255, 0)";
//     else if (value <= 33) return "rgb(132, 255, 0)";
//     else if (value <= 34) return "rgb(142, 255, 0)";
//     else if (value <= 35) return "rgb(153, 255, 0)";
//     else if (value <= 36) return "rgb(163, 255, 0)";
//     else if (value <= 37) return "rgb(173, 255, 0)";
//     else if (value <= 38) return "rgb(183, 255, 0)";
//     else if (value <= 39) return "rgb(193, 255, 0)";
//     else if (value <= 40) return "rgb(255, 255, 0)";
//     else if (value <= 41) return "rgb(255, 245, 0)";
//     else if (value <= 42) return "rgb(255, 235, 0)";
//     else if (value <= 43) return "rgb(255, 225, 0)";
//     else if (value <= 44) return "rgb(255, 215, 0)";
//     else if (value <= 45) return "rgb(255, 204, 0)";
//     else if (value <= 46) return "rgb(255, 194, 0)";
//     else if (value <= 47) return "rgb(255, 184, 0)";
//     else if (value <= 48) return "rgb(255, 174, 0)";
//     else if (value <= 49) return "rgb(255, 164, 0)";
//     else if (value <= 50) return "rgb(255, 153, 0)";
//     else if (value <= 51) return "rgb(255, 143, 0)";
//     else if (value <= 52) return "rgb(255, 133, 0)";
//     else if (value <= 53) return "rgb(255, 123, 0)";
//     else if (value <= 54) return "rgb(255, 113, 0)";
//     else if (value <= 55) return "rgb(255, 102, 0)";
//     else if (value <= 56) return "rgb(255, 77, 0)";
//     else if (value <= 57) return "rgb(255, 51, 0)";
//     else if (value <= 58) return "rgb(255, 26, 0)";
//     else if (value <= 59) return "rgb(255, 13, 0)";
//     else return "rgb(255, 0, 0)";
// }