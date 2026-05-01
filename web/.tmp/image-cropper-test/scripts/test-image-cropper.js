"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const image_cropper_1 = require("../src/utils/image-cropper");
function assertClose(actual, expected, label, epsilon = 0.0001) {
    if (Math.abs(actual - expected) > epsilon) {
        throw new Error(`${label}: expected ${expected}, got ${actual}`);
    }
}
function assertEqual(actual, expected, label) {
    if (actual !== expected) {
        throw new Error(`${label}: expected ${String(expected)}, got ${String(actual)}`);
    }
}
function makeExifOrientationJpeg(orientation) {
    const bytes = [
        0xff, 0xd8,
        0xff, 0xe1, 0x00, 0x22,
        0x45, 0x78, 0x69, 0x66, 0x00, 0x00,
        0x49, 0x49, 0x2a, 0x00,
        0x08, 0x00, 0x00, 0x00,
        0x01, 0x00,
        0x12, 0x01,
        0x03, 0x00,
        0x01, 0x00, 0x00, 0x00,
        orientation, 0x00, 0x00, 0x00,
        0x00, 0x00, 0xff, 0xd9,
    ];
    return Uint8Array.from(bytes).buffer;
}
function testGetMinCropScale() {
    assertClose((0, image_cropper_1.getMinCropScale)(1200, 800, 300), 0.375, 'minScale.landscape');
    assertClose((0, image_cropper_1.getMinCropScale)(800, 1200, 300), 0.375, 'minScale.portrait');
}
function testPinchTransformSupportsShrinkBack() {
    const minScale = (0, image_cropper_1.getMinCropScale)(1200, 800, 300);
    const zoomIn = (0, image_cropper_1.computePinchTransform)({
        startScale: minScale,
        startOffsetX: 20,
        startOffsetY: 40,
        startDistance: 100,
        currentDistance: 180,
        startCenterX: 200,
        startCenterY: 240,
        currentCenterX: 200,
        currentCenterY: 240,
        minScale,
        maxScale: 6,
        sensitivity: 1.18,
    });
    if (zoomIn.scale <= minScale) {
        throw new Error(`pinch.zoomInScale: expected greater than ${minScale}, got ${zoomIn.scale}`);
    }
    const zoomOut = (0, image_cropper_1.computePinchTransform)({
        startScale: zoomIn.scale,
        startOffsetX: zoomIn.offsetX,
        startOffsetY: zoomIn.offsetY,
        startDistance: 180,
        currentDistance: 60,
        startCenterX: 200,
        startCenterY: 240,
        currentCenterX: 200,
        currentCenterY: 240,
        minScale,
        maxScale: 6,
        sensitivity: 1.18,
    });
    assertClose(zoomOut.scale, minScale, 'pinch.zoomOutClamp');
}
function testExtractExifOrientation() {
    assertEqual((0, image_cropper_1.extractExifOrientation)(makeExifOrientationJpeg(3)), 3, 'exif.orientation3');
    assertEqual((0, image_cropper_1.extractExifOrientation)(makeExifOrientationJpeg(6)), 6, 'exif.orientation6');
    assertEqual((0, image_cropper_1.extractExifOrientation)(new Uint8Array([0xff, 0xd8, 0xff, 0xd9]).buffer), 1, 'exif.default');
}
function main() {
    testGetMinCropScale();
    testPinchTransformSupportsShrinkBack();
    testExtractExifOrientation();
    console.log('image-cropper tests passed');
}
main();
