# Voleur
Remote volume mixer (and more)

voleur /vɔlœʀ/ *French*: thief

Use case
-----------
Have you ever started a game while chatting to your friends over VoIP only to have them drowned out by the game audio? Alt-tabbing may break the game, mute the audio, and is just annoying. Using Voleur, you can open the page on your phone and quickly reduce the volume.

Voleur is composed of a JavaScript front-end that displays controls (sliders) and a Go back-end which updates these based on the property they track (such as volume) and sets the property based on the control value. Communication is done using [server-sent events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events) and POST requests.

Voleur is designed to easily allow various "plugins" to be slotted in. The back-end needs only implement `IControlInterface`. Ideas:
* R, G, B channels for an LED strip
* Fan RPM control and temperature indicator

How to build/run
----
There are no releases as I haven't looked into packaging assets yet.

`go get github.com/alexandruioanp/voleur-go`

`cd $GOPATH/src/github.com/alexandruioanp/voleur-go/`

`go run *.go`

Navigate to `<your-ip>:8080` (e.g. `localhost:8080`).

The TypeScript can be transpiled to JavaScript using `tsc index.ts`. A JavaScript version of the front-end code is also provided. 

Supported back-ends
----
* PulseAudio
Voleur currently uses a `pactl`-based PulseAudio back-end by default, which is what the project was started for. It uses `stdin/stdout` to interface with PulseAudio, as I couldn't find suitable bindings for Go.
Icons are retrieved using GTK and its [Python3 bindings](https://python-gtk-3-tutorial.readthedocs.io/en/latest/install.html#dependencies). This can be disabled in `ifaces/pulse_cmdline.go` by setting `include_icon` to `false`.


Roadmap
----
* Drop-down control support
* Additional PulseAudio back-end features (mute button, activity indicator, sink volume, fallback icons, output interface selection)
* Windows volume mixer back-end (WASAPI)

Screenshot
-----
![](https://i.imgur.com/7BWNeGR.png)

Uses
----
[bootstrap-slider](https://github.com/seiyria/bootstrap-slider)

jQuery

Bootstrap
