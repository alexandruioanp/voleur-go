document.body.innerHTML += "HELLO";

interface App {
    name: string;
    volume: number;
    uid: string;
}

var evtSource = new EventSource("/events");

evtSource.onmessage = function(e)
{
    decoded = decode_payload(e);
    // console.log(decoded)
    update_vol(decoded);
}

function decode_payload(stin: string): App[]
{
    return JSON.parse(stin.data) as App[];
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
    // console.log(info.name + " " + info.uid + " not found. Creating...");
    create_div(info);
}

function set_volume(volBox, volume: number)
{
    // console.log("found volbx with ID:" + volBox.id)
    $("#slider" + volBox.id).slider("setValue", volume);
    // console.log("ID:" + volBox.id)
    //console.log("#slider" + volBox.id);
    //console.log($("#slider" + volBox.id).slider);
}

function create_div(info: App)
{
    let volDiv  = document.getElementById('volume-container');
    let sliderElement = make_slider(info);
    let sliderDiv = document.createElement("div");
    sliderDiv.id = info.uid;
    sliderDiv.appendChild(sliderElement);
    volDiv.appendChild(sliderDiv);
    
    $("#" + sliderElement.id).slider({
        reversed : true
    });

    $("#" + sliderElement.id).slider('setValue', String(info.vol));

    $("#" + sliderElement.id).slider().on("change", slider_slid);

    $("#" + sliderElement.id).slider().on("mousedown", slider_click);
}

function slider_click(ev) {
    console.log("Click");
}

function slider_slid(ev)
{
    console.log(ev.target.id + " " + ev.target.value);
    // console.log(JSON.stringify({uid: ev.target.id, vol: ev.target.value}));
    $.post("/volOps", JSON.stringify({uid: ev.target.id.slice("slider".length), vol: parseInt(ev.target.value, 10)}));
}

function make_slider(info: App)
{
    let fakeSlider = document.createElement("div");
//    fakeSlider.innerHTML = '<input type="range" min="1" max="100" value="50">';
    fakeSlider.innerHTML = '<input id="slider" type="text" data-slider-min="0" data-slider-max="100" data-slider-step="1" data-slider-value="-3" data-slider-orientation="vertical"/>';
    let slider = fakeSlider.firstChild;
    // TODO deal with [ and ] somehow 
    slider.id = "slider" + info.uid;
    // console.log("slider id: " + slider.id);

    return slider;
}
