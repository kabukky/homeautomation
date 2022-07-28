var temperatureChart = null;
var calendar = null;
const maxHours = 16; // Only show forecast a few hours into the future
document.fonts.onloadingdone = () => {
    if (temperatureChart) {
        temperatureChart.update();
    } 
};
Chart.defaults.font.family = "Lato";
Chart.defaults.font.weight = "400";
moment.locale("de");

function showCalendar(data) {
    var events = {};
    var colorIndex = 1;

    // We want the order of keys reversed in this case
    // But this is just for me
    // data = reverseObject(data);

    for (var key of Object.keys(data)) {
        // Only supporting two colors for now
        var eventColor = colorIndex % 2 != 0 ? "black" : "white";
        data[key].forEach(function (element) {
            var startDate = new Date(element.start_date);
            var endDate = new Date(element.end_date);
            var event = {
                title: element.name != "" ? element.name: "Beschäftigt",
                color: eventColor,
                startDate: new Date(startDate),
                endDate: new Date(endDate)
            }
            var dates = [new Date(startDate)];
            var days = Math.abs(Math.ceil((startDate.getTime() - endDate.getTime()) / (1000 * 3600 * 24)));
            for (var i = days; i > 0; i--) {
                if (i == 1) {
                    // Last day
                    if (endDate.getHours() == 0 && endDate.getMinutes() == 0 && endDate.getSeconds() == 0) {
                        continue;
                    }
                }
                // Add
                startDate.setDate(startDate.getDate() + 1);
                dates.push(new Date(startDate));
            }
            dates.forEach(function (date) {
                // Check if date is before today
                var today = new Date();
                if (date.getFullYear() < today.getFullYear()) {
                    return;
                }
                if (date.getFullYear() == today.getFullYear() && date.getMonth() < today.getMonth()) {
                    return;
                }
                if (date.getFullYear() == today.getFullYear() && date.getMonth() == today.getMonth() && date.getDate() < today.getDate()) {
                    return;
                }
                
                var key = moment(date).format("YYYY/MM/DD");
                if (!events[key]) {
                    events[key] = []
                }
                events[key].push(event);
            });
        });
        colorIndex++;
    }

    // Add today if not already added
    var keyToday = moment(new Date()).format("YYYY/MM/DD");
    if (!events[keyToday]) {
        events[keyToday] = [];
    }

    // Order events keys
    events = Object.keys(events).sort().reduce(
        (obj, key) => { 
            obj[key] = events[key]; 
            return obj;
        }, 
        {}
    );

    // Display HTML
    var calendarContainers = [document.getElementById("cal-col-1"), document.getElementById("cal-col-2"), document.getElementById("cal-col-3")];
    var containerIndex = 0;
    var rowIndex = 0;
    const maxRowsPerColumn = 7;
    calendarContainers.forEach(function (container) {
        container.innerHTML = '';
    });
    for (var key of Object.keys(events)) {
        // Determine if we should start a new column
        containerIndex = determineColumnIndex(rowIndex, maxRowsPerColumn)
        if (containerIndex == -1) {
            break;
        }
        // We do not want to disyplay the date as the last row in a columnn
        if (rowIndex == maxRowsPerColumn - 1 || rowIndex == (maxRowsPerColumn*2) - 1 || rowIndex == (maxRowsPerColumn*3) - 1) {
            containerIndex++;
            if (containerIndex >= calendarContainers.length) {
                // No more columns available to display events in
                break;
            }
        }
        var keyDate = new Date(key);
        calendarContainers[containerIndex].innerHTML += '<div class="row pt-2 mx-1 date-calendar fw-bold">' + moment(keyDate).format("ddd, Do MMMM YYYY") + '</div>'
        rowIndex++;
        if (events[key].length == 0) {
            containerIndex = determineColumnIndex(rowIndex, maxRowsPerColumn)
            if (containerIndex == -1) {
                return;
            }
            calendarContainers[containerIndex].innerHTML += '<div class="row m-1 p-1"><span class="px-1">Nichts geplant</span></div>'
            rowIndex++;
        } else {
            events[key].forEach(function (event) {
                // Determine if we should start a new column
                containerIndex = determineColumnIndex(rowIndex, maxRowsPerColumn)
                if (containerIndex == -1) {
                    return;
                }
                var startTimeString = moment(event.startDate).format("HH:mm")
                var endTimeString = moment(event.endDate).format("HH:mm")
                if (datesAreOnSameDay(event.startDate, keyDate)) {
                    if (!datesAreOnSameDay(event.endDate, keyDate)) {
                        endTimeString = "";
                        // Event ends on a later day
                        if (startTimeString == "00:00") {
                            startTimeString = "";
                        }
                    }
                } else {
                    // Event starts on an earlier day
                    startTimeString = "";
                }
                if (!datesAreOnSameDay(event.startDate, keyDate) && !datesAreOnSameDay(event.endDate, keyDate)) {
                    // Event runs all day, didn't start start and doesn't end on this day
                    startTimeString = "";
                    endTimeString = "";
                }
                var timeString = startTimeString;
                if (startTimeString != "" && endTimeString != "") {
                    timeString += " - ";
                } else if (startTimeString == "" && endTimeString != "") {
                    timeString = "Bis ";
                }
                timeString += endTimeString;
                var textColorString = "text-white";
                var styleString = 'style="background-color:';
                if (event.color == "black") {
                    styleString += '#000000;"'
                } else {
                    // white
                    textColorString = "text-black";
                    styleString += '#ffffff;border:solid #000000;border-width: 1px;color: #000000"'
                }
                rowIndex++;
                if (timeString != "") {
                    // Additional row for displaying time
                    rowIndex++;
                }
                if (event.title.length > 30) {
                    // Additional row for likely line break in title
                    rowIndex++;
                }
                if (rowIndex >= (maxRowsPerColumn*3)) {
                    // Not enough rows remaing to add this event
                    return;
                }
                calendarContainers[containerIndex].innerHTML += '<div class="row m-1 p-1 rounded ' + textColorString + '" ' + styleString + '><span class="px-1">' + event.title + '</span><small class="px-1">' + timeString + '</small></div>'
            });
        }
    }
}

function showWeather(data) {
    var sunset = new Date(data.current.sunset);
    var sunrise = new Date(data.current.sunrise);

    // Current
    var currentTemperatureContainer = document.getElementById("weather-current-temperature");
    currentTemperatureContainer.innerHTML = data.current.temperature_celsius.toFixed() + '<sup id="main-temp-degrees">°C</sup>';

    var currentTemperatureContainer = document.getElementById("weather-current-icon");
    currentTemperatureContainer.innerHTML = '<i class="wi '+determineIconPrefix(sunset, sunrise, new Date(data.current.time))+data.current.openweathermap_id+'"></i>'

    var ctx = document.getElementById("temperature-chart");
    var chartData = {
        labels: [],
        datasets: [{
            label: "Temperature",
            borderWidth: 3,
            borderColor: "#000000",
            pointBackgroundColor: "#000000",
            pointBorderColor: "#ffffff",
            pointBorderWidth: 3,
            pointRadius: 6,
            datalabels: {
                color: "#000000",
                font: {
                    weight: "700",
                    size: 18,
                },
                formatter: function(value, context) {
                    return Math.round(value) + "°";
                }
            },
            data: []
        },{
            label: "Precipitation",
            borderWidth: 3,
            borderColor: "#000000",
            pointBackgroundColor: "#000000",
            pointBorderColor: "#ffffff",
            pointBorderWidth: 2,
            pointRadius: 4,
            fill: true,
            backgroundColor: "#eeeeee",
            datalabels: {
                color: "#000000",
                font: {
                    weight: "700",
                    size: 13,
                },
                formatter: function(value, context) {
                    if (value == 0) {
                        return "";
                    }
                    return value + " mm";
                }
            },
            data: []
        }]
    };

    // Icons
    var imagesContainer = document.getElementById("temperature-chart-images");
    imagesContainer.innerHTML = '';
    var precipitationProbabilityContainer = document.getElementById("temperature-chart-precipitation-probability");
    precipitationProbabilityContainer.innerHTML = '';

    var hasPrecipitation = false;
    data.forecast.forEach(function (element, index) {
        if (index < maxHours) {
            if (element.precipitation_amount > 0) {
                hasPrecipitation = true;
            }
            var date = new Date(element.time);
            chartData.datasets[0].data.push(element.temperature_celsius);
            chartData.datasets[1].data.push(element.precipitation_amount);
            chartData.labels.push(moment(date).format('HH:mm'));
            imagesContainer.innerHTML += '<div class="col p-0 text-center"><i class="wi '+determineIconPrefix(sunset, sunrise, date)+element.openweathermap_id+'"></i></div>';
            precipitationProbabilityContainer.innerHTML += '<div class="col p-0 text-center"><i class="wi wi-umbrella"></i><br>'+Math.round(element.precipitation_probability*100)+'%</div>';
        }
    });

    if (!hasPrecipitation) {
        chartData.datasets[1].data = [];
    }

    var chartOptions = {
        layout: {
            padding: {
                top: 50
            }
        },
        animation: {
            duration: 0
        },
        elements: {
            line: {
                tension: 0.4
            }
        },
        hover: {
            mode: null
        },
        scales: {
            yAxis: {
                ticks: {
                    display: false
                }
            },
            xAxis: {
                ticks: {
                    color: "#000000",
                    font: {
                        size: 16
                    }
                }
            }
        },
        plugins: {
            datalabels: {
                anchor: "end",
                align: "top",
                offset: 10,
                display: "auto"
            },
            tooltip: {
                enabled: false
            },
            legend: {
                display: false
            }
        }
    };

    if (temperatureChart) {
        temperatureChart.destroy();
    }

    temperatureChart = new Chart(ctx, {
        plugins: [ChartDataLabels],
        type: "line",
        data: chartData,
        options: chartOptions
    });
}

function determineIconPrefix(sunset, sunrise, date) {
    var prefix = "wi-owm-day-";
    if (date.getHours() >= sunset.getHours() || date.getHours() < sunrise.getHours()) {
        prefix = "wi-owm-night-";
    }
    return prefix;
}

function reverseObject(object) {
    var newObject = {};
    var keys = [];

    for (var key in object) {
        keys.push(key);
    }

    for (var i = keys.length - 1; i >= 0; i--) {
        var value = object[keys[i]];
        newObject[keys[i]]= value;
    }       

    return newObject;
}

function datesAreOnSameDay(date1, date2) {
    if (
        date1.getFullYear() === date2.getFullYear() &&
        date1.getMonth() === date2.getMonth() &&
        date1.getDate() === date2.getDate()
    ) {
        return true;
    }
    return false;
}

// determineColumnIndex returns -1 if maximum number of columns is reached (3 columns fow now)
function determineColumnIndex(rowIndex, maxRowsPerColumn) {
    if (rowIndex < maxRowsPerColumn) {
        return 0
    } else if (rowIndex < maxRowsPerColumn*2) {
        return 1
    } else if (rowIndex < maxRowsPerColumn*3) {
        return 2
    } else {
        return - 1;
    }
}

function updateDashboard() {
    // Update date
    var currentDate = moment(new Date());
    var dateContainer = document.getElementById("date");
    dateContainer.innerHTML = currentDate.format("dddd<br>Do MMMM YYYY");
    var timeContainer = document.getElementById("time");
    timeContainer.innerHTML = currentDate.format("HH:mm");

    // Update weather
    fetch("/api/v1/weather")
    .then(response => response.json())
    .then(data => showWeather(data));

    // Update calendar
    fetch("/api/v1/calendar")
    .then(response => response.json())
    .then(data => showCalendar(data));
}

updateDashboard();

t = setInterval(function() {
    updateDashboard();
}, 60*1000); // Every minute
