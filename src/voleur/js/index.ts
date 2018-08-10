interface App {
    name: string;
    vol: number;
    uid: string;
}

var evtSource = new EventSource("/events");

evtSource.onmessage = function(e)
{
    let decoded = decode_payload(e);
    update_vol(decoded);
}

function decode_payload(stin: MessageEvent): App
{
    return JSON.parse(stin.data) as App;
}

function update_vol(info: App)
{
    let volDiv  = document.getElementById('volume-container');
    let volBoxes = volDiv.children;
    for(let i = 0; i < volBoxes.length; i++)
    {
        let box = volBoxes[i];
        if(box.id == info.uid)
        {
            set_volume(box, info.vol);
            return;
        }
    }
    create_div(info);
}

function set_volume(volBox, volume: number)
{
    $("#slider" + volBox.id).slider("setValue", volume);
}

function create_div(info: App)
{
    let volDiv  = document.getElementById('volume-container');
    
    let sliderElement = make_slider(info);
    let sliderDiv = document.createElement("div");
    sliderDiv.id = info.uid;
    sliderDiv.appendChild(sliderElement);
    sliderDiv.classList.add("sliderdiv");
    sliderDiv.classList.add("border")
    sliderDiv.innerHTML += '<p>' + info.name + '</p>';
    
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
}

function slider_mouseup(ev) {
    console.log("mouseup");
}

function slider_slid(ev)
{
    console.log(ev.target.id + " " + ev.target.value);
    $.post("/volOps", JSON.stringify({uid: ev.target.id.slice("slider".length), vol: parseInt(ev.target.value, 10)}));
}

function make_slider(info: App)
{
    let fakeSlider = document.createElement("div");
    fakeSlider.innerHTML += '<input style="float:center;" id="slider" type="text" data-slider-min="0" data-slider-max="100" data-slider-step="1" data-slider-value="-3" data-slider-orientation="vertical"/>';
    let slider = fakeSlider.firstChild;
    slider.id = "slider" + info.uid;
    console.log("slider id: " + slider.id);

    return slider;
}
