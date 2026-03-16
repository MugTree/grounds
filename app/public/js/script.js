function checkLocation(e) {
  e.preventDefault();
  const value = document.getElementById("location-id").value;
  if (value === "") {
    alert("Please select a location");
  } else {
    window.location.assign(e.target.action + value + "/");
  }
}
