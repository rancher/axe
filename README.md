axe
========

## ***Throwing some axes for fun***

![axe](https://media.giphy.com/media/l17uofKSRXJGIsnYNB/giphy.gif)

Whoops! wrong link

Real example is in here

[![asciicast](https://asciinema.org/a/8kUuLp57LDgsjEZe9DcVpzblh.svg)](https://asciinema.org/a/8kUuLp57LDgsjEZe9DcVpzblh)

## Building

`make`

## Running

`./bin/axe --kubeconfig $KUBECONFIG`

## Example

1. Define your root page, shortcuts, viewmap, pageNav, footers, tableEventHandler
2. Run!

```$xslt

drawer = types.Drawer{
	RootPage:  RootPage,
	Shortcuts: Shortcuts,
	ViewMap:   ViewMap,
	PageNav:   PageNav,
	Footers:   Footers,
}
	
app := throwing.NewAppView(clientset, drawer, tableEventHandler)
if err := app.Init(); err != nil {
	return err
}

return app.Run()

```

## License
Copyright (c) 2018 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
