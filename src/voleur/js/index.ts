enum UpdateType {
    AddOrUpdate,
    Remove,
}

interface Update {
    name: string;
    vol: number;
    uid: string;
    type: number;
}

declare var ignore_updates: boolean = false;

var evtSource = new EventSource("/events");

evtSource.onmessage = function(e)
{
    let decoded = decode_payload(e);
    if(decoded.type == UpdateType.AddOrUpdate) {
        console.log("update");
        console.log(decoded);
        update_vol(decoded);
    } else if(decoded.type == UpdateType.Remove) {
        console.log("remove");
        console.log(decoded);
        remove_slider(decoded);
    }    
}

function remove_slider(info: Update) {
    let element = document.getElementById("box" + info.uid);
    console.log("removing");
    console.log(element);
    console.log(element.parentNode);
    element.parentNode.removeChild(element);
}

function decode_payload(stin: MessageEvent): Update
{
    return JSON.parse(stin.data) as Update;
}

function update_vol(info: Update)
{
    let volDiv  = document.getElementById('volume-container');
    let volBoxes = volDiv.children;
    for(let i = 0; i < volBoxes.length; i++)
    {
        let box = volBoxes[i];
        console.log("box id" + box.id);
        if(box.id == "box" + info.uid)
        {
            console.log("found box with ID" + box.id);
            set_volume(box, info);
            return;
        }
    }
    create_div(info);
}

function set_volume(volBox, update: Update)
{
    $("#slider" + update.uid).slider("setValue", update.vol);
}

function create_div(info: Update)
{
    let volDiv  = document.getElementById('volume-container');
    
    let sliderElement = make_slider(info);
    let sliderDiv = document.createElement("div");
    sliderDiv.id = "box" + info.uid;
    sliderDiv.appendChild(sliderElement);
    sliderDiv.classList.add("sliderdiv");
    sliderDiv.classList.add("border")
    sliderDiv.innerHTML += '<p class="app_name">' + info.name + '</p>';
    sliderDiv.innerHTML += '<p class="uid_small">#' + info.uid + '</p>';
    
    volDiv.appendChild(sliderDiv);
    
    $("#" + sliderElement.id).slider({
        reversed : true
    });
    $("#" + sliderElement.id).slider('setValue', String(info.vol));
    $("#" + sliderElement.id).slider().on("change", slider_slid);
    $("#" + sliderDiv.id).on("mousedown", slider_mousedown);
    $("#" + sliderDiv.id).on("mouseup", slider_mouseup);
}

function slider_mousedown(ev) {
    console.log("mousedown");
    let ignore 
}

function slider_mouseup(ev) {
    console.log("mouseup");
}

function slider_slid(ev)
{
    console.log(ev.target.id + " " + ev.target.value);
    $.post("/volOps", JSON.stringify({uid: ev.target.id.slice("slider".length), vol: parseInt(ev.target.value, 10)}));
}

function make_slider(info: Update)
{
    let fakeSlider = document.createElement("div");
    fakeSlider.innerHTML += '<input style="float:center;" id="slider" type="text" data-slider-min="0" data-slider-max="100" data-slider-step="1" data-slider-value="-3" data-slider-orientation="vertical"/>';
    let slider = fakeSlider.firstChild;
    slider.id = "slider" + info.uid;
    console.log("slider id: " + slider.id);

    return slider;
}
