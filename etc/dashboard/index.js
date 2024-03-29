var temperatureChart = null;
var calendar = null;
const maxHours = 16; // Only show forecast a few hours into the future
const temperatureOffset = 8; // Offset for showing temperatures on the chart
document.fonts.onloadingdone = () => {
    if (temperatureChart) {
        temperatureChart.update();
    } 
};
Chart.defaults.font.family = "Lato";
Chart.defaults.font.weight = "400";
moment.locale("de");


function showCameraTimes(data) {
    if (data.error) {
        console.log("error for camera times:", JSON.stringify(data));
        return;
    }
    // Dryer
    var dryerContainer = document.getElementById("dryer-display");
    if (!data.dryer_minutes || data.dryer_minutes == -2) {
        dryerContainer.innerHTML = 'Aus';
    } else if (data.dryer_minutes == -1) {
        dryerContainer.innerHTML = 'Fertig';
    } else {
        dryerContainer.innerHTML = 'Noch ' + convertMinsToString(data.dryer_minutes);
    }
    // Washing machine
    var washingMachineContainer = document.getElementById("washing-machine-display");
    if (!data.washing_machine_minutes || data.washing_machine_minutes == -2) {
        washingMachineContainer.innerHTML = 'Aus';
    } else if (data.washing_machine_minutes == -1) {
        washingMachineContainer.innerHTML = 'Fertig';
    } else {
        washingMachineContainer.innerHTML = 'Noch ' + convertMinsToString(data.washing_machine_minutes);
    }
}

function showCalendar(data) {
    if (data.error) {
        console.log("error for calendar:", JSON.stringify(data));
        return;
    }
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
    calendarContainers.forEach(function (container) {
        container.innerHTML = '';
    });
    var eventsHtml = [];
    for (var key of Object.keys(events)) {
        var keyDate = new Date(key);
        eventsHtml.push({type: 'heading', html: '<div class="row pt-2 mx-1 date-calendar fw-bold">' + moment(keyDate).format("ddd, Do MMMM YYYY") + '</div>'});
        if (events[key].length == 0) {
            eventsHtml.push({type: 'event-empty', html: '<div class="row m-1 p-1"><span class="px-1">Nichts geplant</span></div>'});
        } else {
            // Iterate events
            events[key].forEach(function (event) {
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
                eventsHtml.push({type: 'event', html: '<div class="row m-1 p-1 rounded ' + textColorString + '" ' + styleString + '><span class="px-1">' + event.title + '</span><small class="px-1">' + timeString + '</small></div>'})
            });
        }
    }

    for (var i = 0; i < eventsHtml.length; i++) {
        if (containerIndex >= calendarContainers.length) {
            break;
        }
        // Add to DOM
        var template = document.createElement('template');
        var html = eventsHtml[i].html.trim();
        template.innerHTML = html;
        var elem = template.content.firstChild;
        calendarContainers[containerIndex].appendChild(elem);
        if (!isElementInViewport(elem)) {
            calendarContainers[containerIndex].removeChild(elem);
            // Remove heading if last element was a heading
            var lastIndex = i - 1;
            if (lastIndex >= 0 && eventsHtml[lastIndex].type == 'heading') {
                calendarContainers[containerIndex].removeChild(calendarContainers[containerIndex].lastChild);
                i--;
            }
            i--;
            containerIndex++;
        }
    }
}

function showWeather(data) {
    if (data.error) {
        console.log("error for weather:", JSON.stringify(data));
        return;
    }
    if (!data.forecast) {
        return;
    }
    var sunset = new Date(data.current.sunset);
    var sunrise = new Date(data.current.sunrise);

    // Current
    var currentTemperatureContainer = document.getElementById("weather-current-temperature");
    currentTemperatureContainer.innerHTML = data.current.temperature_celsius.toFixed().replace(/^-0$/, "0") + '<sup id="main-temp-degrees">°C</sup>';

    var currentTemperatureContainer = document.getElementById("weather-current-icon");
    currentTemperatureContainer.innerHTML = '<i class="wi '+determineIconPrefix(sunset, sunrise, new Date(data.current.time))+data.current.openweathermap_id+'"></i>'

    var currentUVIndexContainer = document.getElementById("weather-current-uv-index");
    currentUVIndexContainer.innerHTML = 'UVI&nbsp;&nbsp;<i class="wi wi-day-sunny"></i>&nbsp;'+Math.round(data.current.uv_index);

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
                    return (Math.round(value) - temperatureOffset + "°").replace(/^-0$/, "0");
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
    var uviContainer = document.getElementById("temperature-chart-uv-index");
    uviContainer.innerHTML = '';

    var hasPrecipitation = false;
    var precipitationIndicesToDisplay = [];
    var maxForecastTemperature = -100;
    var minForecastTemperature = 100;
    data.forecast.forEach(function (element, index) {
        if (index < maxHours) {
            if (element.precipitation_amount > 0) {
                hasPrecipitation = true;
            }
            if (element.temperature_celsius > maxForecastTemperature) {
                maxForecastTemperature = element.temperature_celsius
            }
            if (element.temperature_celsius < minForecastTemperature) {
                minForecastTemperature = element.temperature_celsius
            }
            var date = new Date(element.time);
            chartData.datasets[0].data.push(Math.round(element.temperature_celsius) + temperatureOffset);
            chartData.datasets[1].data.push(element.precipitation_amount);

            if (element.precipitation_amount == 0) {
                // Never display 0
            } else if (!precipitationIndicesToDisplay.includes(index-1) && !precipitationIndicesToDisplay.includes(index-2)) {
                // Last two indices were not be displayed
                // Definitely display this one
                precipitationIndicesToDisplay.push(index);
            } else if (index+1 < data.forecast.length && data.forecast[index+1].precipitation_amount > element.precipitation_amount) {
                // Next one is bigger, do not diplay 
            } else if (index > 0 && precipitationIndicesToDisplay.includes(index-1)) {
                // Last one was displayed. Do not display this one
                if (element.precipitation_amount > data.forecast[index-1].precipitation_amount) {
                    // Except this one is bigger, so display this one instead
                    precipitationIndicesToDisplay = precipitationIndicesToDisplay.filter(function(e) { return e != index-1 });
                    precipitationIndicesToDisplay.push(index);
                }
            } else {
                // Display
                precipitationIndicesToDisplay.push(index);
            }

            chartData.labels.push(moment(date).format('HH:mm'));
            imagesContainer.innerHTML += '<div class="col p-0 text-center"><i class="wi '+determineIconPrefix(sunset, sunrise, date)+element.openweathermap_id+'"></i></div>';
            precipitationProbabilityContainer.innerHTML += '<div class="col p-0 text-center"><i class="wi wi-umbrella"></i><br>'+Math.round(element.precipitation_probability*100)+'%</div>';
            var uvIndexString = '<i class="wi wi-day-sunny"></i>'+Math.round(element.uv_index);
            if (Math.round(element.uv_index) < 1) {
                uvIndexString = ""
            }
            uviContainer.innerHTML += '<div class="col p-0 text-center">'+uvIndexString+'</div>';
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
                },
                max: maxForecastTemperature + temperatureOffset + 1,
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
                display: function(context) {
                    if (context.datasetIndex == 0) {
                        return "auto";
                    }
                    if (precipitationIndicesToDisplay.includes(context.dataIndex)) {
                        return true;
                    }
                    return false;
                }
            },
            tooltip: {
                enabled: false
            },
            legend: {
                display: false
            }
        }
    };

    if (!hasPrecipitation) {
        chartOptions.scales.yAxis.min = minForecastTemperature - 1;
    }

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

function isElementInViewport(el) {
    // Special bonus for those using jQuery
    if (typeof jQuery === "function" && el instanceof jQuery) {
        el = el[0];
    }

    var rect = el.getBoundingClientRect();

    return (
        rect.top >= 0 &&
        rect.left >= 0 &&
        rect.bottom <= (window.innerHeight || document.documentElement.clientHeight) && /* or $(window).height() */
        rect.right <= (window.innerWidth || document.documentElement.clientWidth) /* or $(window).width() */
    );
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

function convertMinsToString(minutes) {
    output = "";
    var h = Math.floor(minutes / 60);
    if (h > 0) {
        output += h + " Stunde";
        if (h > 1) {
            output += "n";
        }
    }
    var m = minutes % 60;
    if (m > 0) {
        if (output != "") {
            output += " und "
        }
        output += m + " Minute";
        if (m > 1) {
            output += "n";
        }
    }
    return output;
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
    

    // Update camera times
    fetch("/api/v1/camera")
    .then(response => response.json())
    .then(data => showCameraTimes(data));
}

updateDashboard();

t = setInterval(function() {
    updateDashboard();
}, 60*1000); // Every minute
