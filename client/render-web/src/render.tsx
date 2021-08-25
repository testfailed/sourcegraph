/// <reference types="react/experimental" />
/// <reference types="react-dom/experimental" />

import React from 'react'
import ReactDOMServer from 'react-dom/server'
import { StaticRouter } from 'react-router'

// TODO(sqs): separate into enterprise/oss
import { EnterpriseWebApp } from '@sourcegraph/web/src/enterprise/EnterpriseWebApp'

export interface RenderRequest {
    requestURI: string
    jscontext: object
}

export interface RenderResponse {
    html?: string
    redirectURL?: string
    error?: string
}

export const render = async ({ requestURI, jscontext }: RenderRequest): Promise<RenderResponse> => {
    // TODO(sqs): not parallel-safe
    if (jscontext && Object.keys(jscontext) > 0 /* TODO(sqs): remove this check, just for curl debugging */) {
        global.window.context = jscontext
    }
    global.window.context.PRERENDER = true

    const routerContext: { url?: string } = {}
    const e = (
        // TODO(sqs): wrap in <React.StrictMode>
        <StaticRouter location={requestURI} context={routerContext}>
            <EnterpriseWebApp />
        </StaticRouter>
    )
    // TODO(sqs): figure out how many times to iterate async
    ReactDOMServer.renderToString(e)
    await new Promise(resolve => setTimeout(resolve))
    await new Promise(resolve => setTimeout(resolve))
    await new Promise(resolve => setTimeout(resolve))
    ReactDOMServer.renderToString(e)
    await new Promise(resolve => setTimeout(resolve, 250))
    await new Promise(resolve => setTimeout(resolve))
    ReactDOMServer.renderToString(e)
    await new Promise(resolve => setTimeout(resolve, 50))
    await new Promise(resolve => setTimeout(resolve))
    ReactDOMServer.renderToString(e)
    await new Promise(resolve => setTimeout(resolve, 100))
    await new Promise(resolve => setTimeout(resolve))
    ReactDOMServer.renderToString(e)
    await new Promise(resolve => setTimeout(resolve, 50))
    await new Promise(resolve => setTimeout(resolve))
    ReactDOMServer.renderToString(e)
    await new Promise(resolve => setTimeout(resolve, 50))
    await new Promise(resolve => setTimeout(resolve))
    ReactDOMServer.renderToString(e)
    await new Promise(resolve => setTimeout(resolve, 50))
    await new Promise(resolve => setTimeout(resolve))
    const html = ReactDOMServer.renderToString(e)

    return {
        html,
        redirectURL: routerContext.url,
    }
}
