const EventListener = { addEventListener: () => null, removeEventListener: () => null }
global.window = {
    navigator: { platform: '', userAgent: '' },
    location: {},
    dispatchEvent: () => {},
    context: require('./jscontext').JSCONTEXT,
    matchMedia: () => ({ matches: false, ...EventListener }),
    ...EventListener,
}
global.localStorage = {
    getItem: () => null,
    setItem: () => {},
    removeItem: () => {},
}
global.document = {
    querySelector: () => null,
    createEvent: () => ({ initCustomEvent: () => {}, ...EventListener }),
    documentElement: {
        setAttribute: () => {},
        classList: {
            add: () => {},
        },
        style: {},
    },
    createElement: () => null,
    ...EventListener,
}
global.Node = {}
global.navigator = window.navigator
global.location = window.location
global.Element = {
    scroll: null,
    prototype: {
        matches: () => false,
        scroll: () => {},
    },
}
window.Element = global.Element
global.self = { ...window, close: () => {} }
global.fetch = require('cross-fetch')
