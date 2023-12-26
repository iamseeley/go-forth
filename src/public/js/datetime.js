
function updateDateTime() {
    var now = new Date();
    var dateString = now.toDateString();
    var timeString = now.toLocaleTimeString();
    document.getElementById("currentDateTime").textContent = dateString + " " + timeString;
}

document.addEventListener('DOMContentLoaded', (event) => {
    updateDateTime();
    setInterval(updateDateTime, 1000);
});
