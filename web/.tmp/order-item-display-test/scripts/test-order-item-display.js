"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const order_item_display_1 = require("../src/utils/order-item-display");
function assertEqual(actual, expected, label) {
    if (actual !== expected) {
        throw new Error(`${label}: expected "${String(expected)}", got "${String(actual)}"`);
    }
}
function testKeepsRetailProductNameWithSpec() {
    const actual = (0, order_item_display_1.getReceiptItemDisplayName)('奈夫·精选五折冻干系列 · 两包', true);
    assertEqual(actual, '奈夫·精选五折冻干系列 · 两包', 'retail.direct.spec');
}
function testStripsPetPrefixForServiceItems() {
    const actual = (0, order_item_display_1.getReceiptItemDisplayName)('树街 · 精洗', false);
    assertEqual(actual, '精洗', 'service.pet.prefix');
}
function testStripsRetailPetPrefixWhenPrefixIsPresent() {
    const actual = (0, order_item_display_1.getReceiptItemDisplayName)('树街 · 奈夫冻干(两包)', true, ['树街']);
    assertEqual(actual, '奈夫冻干(两包)', 'retail.pet.prefix');
}
function testStripsRetailPetPrefixWithoutParens() {
    const actual = (0, order_item_display_1.getReceiptItemDisplayName)('憨憨 · Vetex澳洲生物素·正品保障', true, ['奶球', '憨憨']);
    assertEqual(actual, 'Vetex澳洲生物素·正品保障', 'retail.pet.prefix.no.parens');
}
function testStripsLegacyRetailGroupPrefix() {
    const actual = (0, order_item_display_1.getReceiptItemDisplayName)('零售商品 · 球球 · 奈夫·奶芙豆系列', true, ['球球']);
    assertEqual(actual, '奈夫·奶芙豆系列', 'retail.group.prefix');
}
function testStripsBoardingStaySuffixFromReceiptName() {
    const actual = (0, order_item_display_1.getReceiptItemDisplayName)('房间1 · 康娜温柔乡 · 寄养住宿', false, [], 'boarding');
    assertEqual(actual, '康娜温柔乡', 'boarding.receipt.item.name');
}
function testStripsBoardingStaySuffixFromSpaceSeparatedReceiptName() {
    const actual = (0, order_item_display_1.getReceiptItemDisplayName)('房间1 · 康娜温柔乡 寄养住宿', false, [], 'boarding');
    assertEqual(actual, '康娜温柔乡', 'boarding.receipt.item.name.space.separated');
}
function testMapsBoardingReceiptGroupName() {
    const actual = (0, order_item_display_1.getReceiptGroupName)('房间1', 'boarding');
    assertEqual(actual, '寄养托管', 'boarding.receipt.group.name');
}
function main() {
    testKeepsRetailProductNameWithSpec();
    testStripsPetPrefixForServiceItems();
    testStripsRetailPetPrefixWhenPrefixIsPresent();
    testStripsRetailPetPrefixWithoutParens();
    testStripsLegacyRetailGroupPrefix();
    testStripsBoardingStaySuffixFromReceiptName();
    testStripsBoardingStaySuffixFromSpaceSeparatedReceiptName();
    testMapsBoardingReceiptGroupName();
    console.log('order-item-display tests passed');
}
main();
