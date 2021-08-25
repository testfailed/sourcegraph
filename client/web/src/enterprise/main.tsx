// This is the entry point for the enterprise web app

// Order is important here
// Don't remove the empty lines between these imports

import '@sourcegraph/shared/src/polyfills'

import '../sentry'

import React from 'react'
import ReactDOM from 'react-dom'
import { BrowserRouter } from 'react-router-dom'

import { EnterpriseWebApp } from './EnterpriseWebApp'

// It's important to have a root component in a separate file to create a react-refresh boundary and avoid page reload.
// https://github.com/pmmmwh/react-refresh-webpack-plugin/blob/main/docs/TROUBLESHOOTING.md#edits-always-lead-to-full-reload
// window.addEventListener('DOMContentLoaded', () => {
const root = document.querySelector('#root')!
// TODO(sqs): <React.StrictMode> causes many problems currently, fix those! then wrap the app in <React.StrictMode>.

const jsx = (
    <BrowserRouter>
        <EnterpriseWebApp />
    </BrowserRouter>
)

const hydrate = root.hasChildNodes()
if (hydrate) {
    ReactDOM.hydrateRoot(root, jsx)
} else {
    ReactDOM.createRoot(root).render(jsx)
}
// })
