"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.createCropperPreviewUrl = exports.extractExifOrientation = exports.computePinchTransform = exports.getMinCropScale = exports.clampNumber = void 0;
const DEFAULT_PINCH_SENSITIVITY = 1.18;
const DEFAULT_EXIF_ORIENTATION = 1;
const CROPPER_PREVIEW_MAX_SIDE = 2400;
const CROPPER_PREVIEW_QUALITY = 0.92;
function clampNumber(value, min, max) {
    return Math.min(max, Math.max(min, value));
}
exports.clampNumber = clampNumber;
function getMinCropScale(imageWidth, imageHeight, cropSize) {
    if (imageWidth <= 0 || imageHeight <= 0 || cropSize <= 0) {
        return 1;
    }
    return Math.max(cropSize / imageWidth, cropSize / imageHeight);
}
exports.getMinCropScale = getMinCropScale;
function computePinchTransform(options) {
    const { startScale, startOffsetX, startOffsetY, startDistance, currentDistance, startCenterX, startCenterY, currentCenterX, currentCenterY, minScale, maxScale, sensitivity = DEFAULT_PINCH_SENSITIVITY, } = options;
    const safeStartScale = Math.max(startScale, 0.0001);
    const safeStartDistance = Math.max(startDistance, 0.0001);
    const distanceRatio = Math.max(currentDistance, 0.0001) / safeStartDistance;
    const nextScale = clampNumber(safeStartScale * Math.pow(distanceRatio, sensitivity), Math.max(minScale, 0.0001), Math.max(maxScale, minScale));
    const anchorImageX = (startCenterX - startOffsetX) / safeStartScale;
    const anchorImageY = (startCenterY - startOffsetY) / safeStartScale;
    return {
        scale: nextScale,
        offsetX: currentCenterX - anchorImageX * nextScale,
        offsetY: currentCenterY - anchorImageY * nextScale,
    };
}
exports.computePinchTransform = computePinchTransform;
function extractExifOrientation(buffer) {
    const view = new DataView(buffer);
    if (view.byteLength < 4 || view.getUint16(0, false) !== 0xffd8) {
        return DEFAULT_EXIF_ORIENTATION;
    }
    let offset = 2;
    while (offset + 4 <= view.byteLength) {
        if (view.getUint8(offset) !== 0xff) {
            break;
        }
        const marker = view.getUint8(offset + 1);
        if (marker === 0xda || marker === 0xd9) {
            break;
        }
        const segmentLength = view.getUint16(offset + 2, false);
        if (segmentLength < 2 || offset + 2 + segmentLength > view.byteLength) {
            break;
        }
        if (marker === 0xe1 && segmentLength >= 8) {
            const exifOffset = offset + 4;
            if (view.getUint32(exifOffset, false) === 0x45786966 &&
                view.getUint16(exifOffset + 4, false) === 0x0000) {
                const tiffOffset = exifOffset + 6;
                if (tiffOffset + 8 > view.byteLength) {
                    return DEFAULT_EXIF_ORIENTATION;
                }
                const byteOrder = view.getUint16(tiffOffset, false);
                const littleEndian = byteOrder === 0x4949;
                if (!littleEndian && byteOrder !== 0x4d4d) {
                    return DEFAULT_EXIF_ORIENTATION;
                }
                const firstIfdOffset = view.getUint32(tiffOffset + 4, littleEndian);
                const ifdOffset = tiffOffset + firstIfdOffset;
                if (ifdOffset + 2 > view.byteLength) {
                    return DEFAULT_EXIF_ORIENTATION;
                }
                const entryCount = view.getUint16(ifdOffset, littleEndian);
                for (let index = 0; index < entryCount; index += 1) {
                    const entryOffset = ifdOffset + 2 + index * 12;
                    if (entryOffset + 12 > view.byteLength) {
                        break;
                    }
                    if (view.getUint16(entryOffset, littleEndian) === 0x0112) {
                        return view.getUint16(entryOffset + 8, littleEndian);
                    }
                }
            }
        }
        offset += 2 + segmentLength;
    }
    return DEFAULT_EXIF_ORIENTATION;
}
exports.extractExifOrientation = extractExifOrientation;
async function createCropperPreviewUrl(file) {
    const normalizedBlob = await createCropperPreviewBlob(file);
    return URL.createObjectURL(normalizedBlob || file);
}
exports.createCropperPreviewUrl = createCropperPreviewUrl;
async function createCropperPreviewBlob(file) {
    if (typeof window === 'undefined' || typeof document === 'undefined') {
        return null;
    }
    const orientation = await readExifOrientation(file);
    const shouldReencode = orientation !== DEFAULT_EXIF_ORIENTATION || isHeicLikeFile(file) || file.size > 2 * 1024 * 1024;
    if (!shouldReencode) {
        return null;
    }
    const bitmap = await loadImageBitmap(file, 'from-image');
    if (bitmap) {
        try {
            return await renderPreviewBlob({
                source: bitmap,
                sourceWidth: bitmap.width,
                sourceHeight: bitmap.height,
                outputOrientation: DEFAULT_EXIF_ORIENTATION,
            });
        }
        finally {
            bitmap.close();
        }
    }
    if (orientation === DEFAULT_EXIF_ORIENTATION) {
        return null;
    }
    const image = await loadHtmlImage(file);
    return renderPreviewBlob({
        source: image,
        sourceWidth: image.naturalWidth || image.width,
        sourceHeight: image.naturalHeight || image.height,
        outputOrientation: orientation,
    });
}
async function readExifOrientation(file) {
    if (!looksLikeJpeg(file)) {
        return DEFAULT_EXIF_ORIENTATION;
    }
    const buffer = await file.arrayBuffer();
    return extractExifOrientation(buffer);
}
function looksLikeJpeg(file) {
    const lowerName = file.name.toLowerCase();
    return file.type === 'image/jpeg' || file.type === 'image/jpg' || lowerName.endsWith('.jpg') || lowerName.endsWith('.jpeg');
}
function isHeicLikeFile(file) {
    const lowerName = file.name.toLowerCase();
    return lowerName.endsWith('.heic') || lowerName.endsWith('.heif') || file.type.includes('heic') || file.type.includes('heif');
}
async function loadImageBitmap(file, imageOrientation) {
    if (typeof createImageBitmap !== 'function') {
        return null;
    }
    try {
        return await createImageBitmap(file, { imageOrientation });
    }
    catch {
        return null;
    }
}
async function loadHtmlImage(file) {
    const objectUrl = URL.createObjectURL(file);
    try {
        return await new Promise((resolve, reject) => {
            const image = new Image();
            image.onload = () => resolve(image);
            image.onerror = () => reject(new Error('图片读取失败'));
            image.src = objectUrl;
        });
    }
    finally {
        URL.revokeObjectURL(objectUrl);
    }
}
async function renderPreviewBlob(options) {
    const { source, sourceWidth, sourceHeight, outputOrientation } = options;
    const outputSize = getOrientedSize(sourceWidth, sourceHeight, outputOrientation);
    const ratio = getFitRatio(outputSize.width, outputSize.height, CROPPER_PREVIEW_MAX_SIDE);
    const canvasWidth = Math.max(1, Math.round(outputSize.width * ratio));
    const canvasHeight = Math.max(1, Math.round(outputSize.height * ratio));
    const drawWidth = Math.max(1, Math.round(sourceWidth * ratio));
    const drawHeight = Math.max(1, Math.round(sourceHeight * ratio));
    const canvas = document.createElement('canvas');
    canvas.width = canvasWidth;
    canvas.height = canvasHeight;
    const ctx = canvas.getContext('2d');
    if (!ctx) {
        throw new Error('浏览器不支持图片处理');
    }
    ctx.fillStyle = '#FFFFFF';
    ctx.fillRect(0, 0, canvasWidth, canvasHeight);
    ctx.imageSmoothingEnabled = true;
    ctx.imageSmoothingQuality = 'high';
    applyExifTransform(ctx, outputOrientation, canvasWidth, canvasHeight);
    ctx.drawImage(source, 0, 0, drawWidth, drawHeight);
    return new Promise((resolve, reject) => {
        canvas.toBlob((blob) => {
            if (!blob) {
                reject(new Error('图片转换失败'));
                return;
            }
            resolve(blob);
        }, 'image/jpeg', CROPPER_PREVIEW_QUALITY);
    });
}
function getFitRatio(width, height, maxSide) {
    if (width <= maxSide && height <= maxSide) {
        return 1;
    }
    return width > height ? maxSide / width : maxSide / height;
}
function getOrientedSize(width, height, orientation) {
    if ([5, 6, 7, 8].includes(orientation)) {
        return { width: height, height: width };
    }
    return { width, height };
}
function applyExifTransform(ctx, orientation, width, height) {
    switch (orientation) {
        case 2:
            ctx.transform(-1, 0, 0, 1, width, 0);
            break;
        case 3:
            ctx.transform(-1, 0, 0, -1, width, height);
            break;
        case 4:
            ctx.transform(1, 0, 0, -1, 0, height);
            break;
        case 5:
            ctx.transform(0, 1, 1, 0, 0, 0);
            break;
        case 6:
            ctx.transform(0, 1, -1, 0, width, 0);
            break;
        case 7:
            ctx.transform(0, -1, -1, 0, width, height);
            break;
        case 8:
            ctx.transform(0, -1, 1, 0, 0, height);
            break;
        default:
            break;
    }
}
