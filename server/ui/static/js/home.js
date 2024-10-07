//
// process data
//
// 'data' variable is loaded with data fetched from the server
const dataProcessed = data.map((element) => {
    // convert timestamp from server to locale string
    const timeOfCalculation = new Date(element.time_stamp_calculation)
    const timeString = timeOfCalculation.toLocaleString()

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
// map
//
// map centered on India
let map = L.map("map").setView([20, 78], 5);

// add a title layer with the Open Street Map information
L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
}).addTo(map);


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
        const color = getColor(temperature_wet_bulb);

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
function getColor(temperature) {
    // ensure temperature values are between -100 and 100
    const value = Math.min(100, Math.max(-100, temperature));

    // classify colour according to temperature values
    if (value <= 0) {
        return "rgb(0, 0, 255)"; // blue [-100, 0]
    } else if (value <= 5) {
        return "rgb(0, 191, 255)"; // sky-blue (0, 5]
    } else if (value <= 10) {
        return "rgb(0, 255, 255)"; // cyan (5, 10]
    } else if (value <= 15) {
        return "rgb(0, 255, 127)"; // light-green (10, 15]
    } else if (value <= 20) {
        return "rgb(0, 255, 0)"; // green (15, 20]
    } else if (value <= 25) {
        return "rgb(173, 255, 47)"; // yellow-green (20, 25]
    } else if (value <= 30) {
        return "rgb(255, 255, 0)"; // yellow (25, 30]
    } else if (value <= 35) {
        return "rgb(255, 140, 0)"; // yellow-orange (30, 35]
    } else if (value <= 40) {
        return "rgb(255, 69, 0)"; // orange (35, 40]
    } else {
        return "rgb(255, 0, 0)"; // red (40, 100]
    }
}