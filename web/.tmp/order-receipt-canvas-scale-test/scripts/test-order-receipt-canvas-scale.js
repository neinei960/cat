"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const strict_1 = __importDefault(require("node:assert/strict"));
const receipt_image_1 = require("../src/utils/receipt-image");
function main() {
    strict_1.default.equal((0, receipt_image_1.getReceiptCanvasScale)(365, 823, false), 2, 'short desktop receipts should keep the crisp 2x scale');
    const tallMobileScale = (0, receipt_image_1.getReceiptCanvasScale)(390, 5200, true);
    (0, strict_1.default)(tallMobileScale < 2, 'tall mobile receipts should reduce scale to avoid iOS canvas limits');
    (0, strict_1.default)(tallMobileScale >= 1, 'mobile fallback should keep readable output');
    const boundedPixels = 390 * 5200 * tallMobileScale * tallMobileScale;
    (0, strict_1.default)(boundedPixels <= 4000000, 'mobile receipt canvas should stay within the conservative iOS pixel budget');
}
main();
console.log('order receipt canvas scale checks passed');
