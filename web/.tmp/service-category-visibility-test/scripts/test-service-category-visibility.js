"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const service_category_visibility_1 = require("../src/utils/service-category-visibility");
const categories = [
    {
        ID: 1,
        name: '猫咪洗浴',
        children: [
            { ID: 11, name: '护理类' },
        ],
    },
    {
        ID: 6,
        name: '寄养托管',
        children: [
            { ID: 61, name: '洗浴附加' },
        ],
    },
    {
        ID: 9,
        name: '上门喂养',
    },
];
const visible = (0, service_category_visibility_1.filterServiceManagementCategoryTree)(categories);
if (visible.length !== 1 || visible[0].ID !== 1) {
    throw new Error(`expected only grooming category, got ${visible.map((c) => c.name).join(',')}`);
}
const hiddenIds = (0, service_category_visibility_1.getHiddenServiceManagementCategoryIds)(categories);
for (const id of [6, 61, 9]) {
    if (!hiddenIds.has(id)) {
        throw new Error(`expected hidden category id ${id}`);
    }
}
if (hiddenIds.has(1) || hiddenIds.has(11)) {
    throw new Error('expected grooming category ids to remain visible');
}
