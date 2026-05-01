"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.getReceiptItemDisplayName = exports.getReceiptGroupName = exports.splitOrderItemName = void 0;
function splitOrderItemSegments(name) {
    return String(name || '')
        .split(' · ')
        .map((part) => part.trim())
        .filter(Boolean);
}
function splitOrderItemName(name) {
    const parts = splitOrderItemSegments(name);
    if (parts.length < 2)
        return ['', String(name || '')];
    return [parts[0], parts.slice(1).join(' · ')];
}
exports.splitOrderItemName = splitOrderItemName;
function stripBoardingReceiptSuffix(name) {
    return name.replace(/(?:\s*·\s*|\s+)寄养住宿$/, '').trim();
}
function getReceiptGroupName(groupName, orderKind) {
    const normalizedGroupName = String(groupName || '').trim();
    if (!normalizedGroupName)
        return '';
    if (orderKind === 'boarding')
        return '寄养托管';
    return normalizedGroupName;
}
exports.getReceiptGroupName = getReceiptGroupName;
function getReceiptItemDisplayName(name, isRetailGroup, retailPrefixes = [], orderKind) {
    const rawName = String(name || '').trim();
    if (!rawName)
        return '';
    if (!isRetailGroup) {
        const [, itemName] = splitOrderItemName(rawName);
        if (orderKind === 'boarding') {
            return stripBoardingReceiptSuffix(itemName);
        }
        return itemName;
    }
    const parts = splitOrderItemSegments(rawName);
    if (parts.length <= 1)
        return rawName;
    const normalizedPrefixes = new Set(['零售商品', ...retailPrefixes]
        .map((part) => String(part || '').trim())
        .filter(Boolean));
    let startIndex = 0;
    while (startIndex < parts.length - 1 && normalizedPrefixes.has(parts[startIndex])) {
        startIndex += 1;
    }
    if (startIndex > 0) {
        return parts.slice(startIndex).join(' · ');
    }
    // Backward-compatible fallback for legacy "宠物名 · 商品名(规格)" retail rows.
    const secondPart = parts[1];
    if (secondPart.includes('(') || secondPart.includes('（')) {
        return secondPart;
    }
    return rawName;
}
exports.getReceiptItemDisplayName = getReceiptItemDisplayName;
