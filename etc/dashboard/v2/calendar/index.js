var calendar = null;
moment.locale("de");


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
        eventsHtml.push({type: 'heading', html: '<div class="row pt-2 date-calendar fw-bold">' + moment(keyDate).format("ddd, Do MMMM YYYY") + '</div>'});
        if (events[key].length == 0) {
            eventsHtml.push({type: 'event-empty', html: '<div class="row my-1 p-1 calendar-none"><span class="px-1">Nichts geplant</span></div>'});
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
                    styleString += '#00FF00;"'
                } else {
                    // white
                    textColorString = "text-white";
                    styleString += '#FF0000;"'
                }
                eventsHtml.push({type: 'event', html: '<div class="entry-calendar row my-1 p-1 rounded ' + textColorString + '" ' + styleString + '><span class="px-1">' + event.title + '</span><small class="px-1">' + timeString + '</small></div>'})
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

function updateDashboard() {
    // Update calendar
    fetch("/api/v1/calendar")
    .then(response => response.json())
    .then(data => showCalendar(data));
}

const img = new Image();
img.style.maxHeight = "1200px";
img.style.width= "auto";

img.onload = function() {
    const width = img.naturalWidth;
    const height = img.naturalHeight;

    // Calculate the aspect ratio
    const aspectRatio = width / height;

    if (aspectRatio > 1) {
        imageContainer = document.getElementById('image-landscape');
        imageContainer.appendChild(img);
    } else {
        imageContainer = document.getElementById('image-portrait');
        imageContainer.appendChild(img);
    }
    updateDashboard();
};

// Set the source to trigger loading
img.src = "/api/v1/picture";
