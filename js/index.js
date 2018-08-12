var UpdateType;
(function (UpdateType) {
    UpdateType[UpdateType["AddOrUpdate"] = 0] = "AddOrUpdate";
    UpdateType[UpdateType["Remove"] = 1] = "Remove";
})(UpdateType || (UpdateType = {}));
var ignore_updates = false;
var ignore_updates_enabled = true;
var ignored_uid = "";
var evtSource = new EventSource("/events");
evtSource.onmessage = function (e) {
    var decoded = decode_payload(e);
    if (decoded.type == UpdateType.AddOrUpdate) {
        update_val(decoded);
    }
    else if (decoded.type == UpdateType.Remove) {
        remove_slider(decoded);
    }
};
function remove_slider(info) {
    var element = document.getElementById("box" + info.uid);
    element.parentNode.removeChild(element);
}
function decode_payload(stin) {
    return JSON.parse(stin.data);
}
function update_val(info) {
    // console.log("Ignore? " + ignore_updates + " " + ignored_uid == info.uid)
    if (ignore_updates_enabled && ignore_updates && ignored_uid == info.uid) {
        return;
    }
    var box = document.getElementById("box" + info.uid);
    if (box) {
        set_value(box, info);
    }
    else {
        create_div(info);
    }
}
function set_value(valBox, update) {
    $("#slider" + update.uid).slider("setValue", update.val);
}
function create_div(info) {
    var valDiv = document.getElementById('value-container');
    var sliderElement = make_slider(info);
    var sliderDiv = document.createElement("div");
    sliderDiv.id = "box" + info.uid;
    sliderDiv.appendChild(sliderElement);
    sliderDiv.classList.add("sliderdiv");
    sliderDiv.classList.add("border");
    new_html = '<div class="slider-info-container">';
    new_html += '<p class="app_name">' + info.name + '</p>';
    new_html += '<p class="uid_small">#' + info.uid + '</p>';
    new_html += '</div>';
    sliderDiv.innerHTML += new_html;
    if (info.auxdata.icon) {
        var image = new Image();
        // https://stackoverflow.com/questions/27886677/javascript-get-extension-from-base64-image
        var decodedData = window.atob(info.auxdata.icon.slice(0, 20));
        var extension = undefined;
        // do something like this
        var lowerCase = decodedData.toLowerCase();
        // console.log(lowerCase);
        if (lowerCase.indexOf("png") !== -1) {
            extension = "png";
        }
        else if (lowerCase.indexOf("jpg") !== -1 || lowerCase.indexOf("jpeg") !== -1) {
            extension = "jpg";
        }
        else if (lowerCase.indexOf("svg") !== -1 || lowerCase.indexOf("xml") !== -1) {
            extension = "svg+xml";
        }
        else {
            extension = "tiff";
        }
        // alternatively, you can do this
        image.src = "data:image/" + extension + ";base64," + info.auxdata.icon;
        sliderDiv.appendChild(image);
    }
    valDiv.appendChild(sliderDiv);
    console.log(info);
    $("#" + sliderElement.id).slider({
        reversed: true
    });
    $("#" + sliderElement.id).slider('setValue', String(info.val));
    $("#" + sliderElement.id).slider().on("change", slider_slid);
    $("#" + sliderDiv.id).on("mousedown", slider_mousedown);
    $("#" + sliderDiv.id).on("mouseup", slider_mouseup);
}
function slider_mousedown(ev) {
    // console.log("mousedown");
    // console.log(ev.currentTarget.id);
    ignore_updates = true;
    ignored_uid = ev.currentTarget.id.slice("box".length);
}
function slider_mouseup(ev) {
    // console.log("mouseup");
    // console.log(ev.currentTarget.id);
    ignore_updates = false;
}
function slider_slid(ev) {
    console.log(ev.target.id + " " + ev.target.value);
    $.post("/valOps", JSON.stringify({ uid: ev.target.id.slice("slider".length), val: parseInt(ev.target.value, 10) }));
}
function make_slider(info) {
    var fakeSlider = document.createElement("div");
    fakeSlider.innerHTML += '<input style="float:center;" id="slider" type="text" data-slider-min="0" data-slider-max="100" data-slider-step="1" data-slider-value="-3" data-slider-orientation="vertical"/>';
    var slider = fakeSlider.firstChild;
    slider.id = "slider" + info.uid;
    return slider;
}
