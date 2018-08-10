var evtSource = new EventSource("/events");
evtSource.onmessage = function (e) {
    var decoded = decode_payload(e);
    update_vol(decoded);
};
function decode_payload(stin) {
    return JSON.parse(stin.data);
}
function update_vol(info) {
    var volDiv = document.getElementById('volume-container');
    var volBoxes = volDiv.children;
    for (var i = 0; i < volBoxes.length; i++) {
        var box = volBoxes[i];
        if (box.id == info.uid) {
            set_volume(box, info.vol);
            return;
        }
    }
    create_div(info);
}
function set_volume(volBox, volume) {
    $("#slider" + volBox.id).slider("setValue", volume);
}
function create_div(info) {
    var volDiv = document.getElementById('volume-container');
    var sliderElement = make_slider(info);
    var sliderDiv = document.createElement("div");
    sliderDiv.id = info.uid;
    sliderDiv.appendChild(sliderElement);
    sliderDiv.classList.add("sliderdiv");
    sliderDiv.classList.add("border");
    sliderDiv.innerHTML += '<p>' + info.name + '</p>';
    volDiv.appendChild(sliderDiv);
    $("#" + sliderElement.id).slider({
        reversed: true
    });
    $("#" + sliderElement.id).slider('setValue', String(info.vol));
    $("#" + sliderElement.id).slider().on("change", slider_slid);
    console.log($("#" + sliderElement.id));
    $("#" + sliderDiv.id).on("mousedown", slider_mousedown);
    $("#" + sliderDiv.id).on("mouseup", slider_mouseup);
}
function slider_mousedown(ev) {
    console.log("mousedown");
}
function slider_mouseup(ev) {
    console.log("mouseup");
}
function slider_slid(ev) {
    console.log(ev.target.id + " " + ev.target.value);
    $.post("/volOps", JSON.stringify({ uid: ev.target.id.slice("slider".length), vol: parseInt(ev.target.value, 10) }));
}
function make_slider(info) {
    var fakeSlider = document.createElement("div");
    fakeSlider.innerHTML += '<input style="float:center;" id="slider" type="text" data-slider-min="0" data-slider-max="100" data-slider-step="1" data-slider-value="-3" data-slider-orientation="vertical"/>';
    var slider = fakeSlider.firstChild;
    slider.id = "slider" + info.uid;
    console.log("slider id: " + slider.id);
    return slider;
}
