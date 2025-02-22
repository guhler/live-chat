(function() {
    /** @type {import("../htmx").HtmxInternalApi} */
    var api

    var attrPrefix = 'hx-swap-'

    // IE11 doesn't support string.startsWith
    function startsWith(str, prefix) {
        return str.substring(0, prefix.length) === prefix
    }

    /**
       * @param {HTMLElement} elt
       * @param {number} respCode
       * @returns {HTMLElement | null}
       */
    function getRespCodeSwap(elt, respCodeNumber) {
        if (!elt || !respCodeNumber) return null

        var respCode = respCodeNumber.toString()

        // '*' is the original syntax, as the obvious character for a wildcard.
        // The 'x' alternative was added for maximum compatibility with HTML
        // templating engines, due to ambiguity around which characters are
        // supported in HTML attributes.
        //
        // Start with the most specific possible attribute and generalize from
        // there.
        var attrPossibilities = [
            respCode,

            respCode.substring(0, 2) + '*',
            respCode.substring(0, 2) + 'x',

            respCode.substring(0, 1) + '*',
            respCode.substring(0, 1) + 'x',
            respCode.substring(0, 1) + '**',
            respCode.substring(0, 1) + 'xx',

            '*',
            'x',
            '***',
            'xxx'
        ]
        if (startsWith(respCode, '4') || startsWith(respCode, '5')) {
            attrPossibilities.push('error')
        }

        for (var i = 0; i < attrPossibilities.length; i++) {
            var attr = attrPrefix + attrPossibilities[i]
            var attrValue = api.getClosestAttributeValue(elt, attr)
            if (attrValue) {
                return attrValue
            }
        }

        return null
    }

    htmx.defineExtension('response-swap', {

        /** @param {import("../htmx").HtmxInternalApi} apiRef */
        init: function(apiRef) {
            api = apiRef
        },

        /**
             * @param {string} name
             * @param {Event} evt
             */
        onEvent: function(name, evt) {
            if (name === 'htmx:beforeSwap' &&
                evt.detail.xhr &&
                evt.detail.xhr.status !== 200) {
                if (!evt.detail.requestConfig) {
                    return true
                }
                var swap = getRespCodeSwap(evt.detail.requestConfig.elt, evt.detail.xhr.status)
                if (swap) {
                    evt.detail.shouldSwap = true
                    evt.detail.swapOverride = swap
                }
                return true
            }
        }
    })
})()
