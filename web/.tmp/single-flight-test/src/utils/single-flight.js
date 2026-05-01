"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.singleFlight = void 0;
function singleFlight(action) {
    let inFlight = false;
    return async (...args) => {
        if (inFlight)
            return undefined;
        inFlight = true;
        try {
            return await action(...args);
        }
        finally {
            inFlight = false;
        }
    };
}
exports.singleFlight = singleFlight;
