"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.getReceiptCanvasScale = void 0;
const DESKTOP_RECEIPT_SCALE = 2;
const MOBILE_RECEIPT_MAX_PIXELS = 4000000;
function getReceiptCanvasScale(width, height, constrainedMobile) {
    const safeWidth = Math.max(1, Math.ceil(width || 0));
    const safeHeight = Math.max(1, Math.ceil(height || 0));
    if (!constrainedMobile)
        return DESKTOP_RECEIPT_SCALE;
    const maxScale = Math.sqrt(MOBILE_RECEIPT_MAX_PIXELS / (safeWidth * safeHeight));
    return Math.max(1, Math.min(DESKTOP_RECEIPT_SCALE, Number(maxScale.toFixed(2))));
}
exports.getReceiptCanvasScale = getReceiptCanvasScale;
