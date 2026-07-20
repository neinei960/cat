"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.getHiddenServiceManagementCategoryIds = exports.filterServiceManagementCategoryTree = exports.HIDDEN_SERVICE_MANAGEMENT_CATEGORY_NAMES = void 0;
exports.HIDDEN_SERVICE_MANAGEMENT_CATEGORY_NAMES = new Set(['寄养托管', '上门喂养']);
function normalizeCategoryName(name) {
    return String(name || '').trim();
}
function isHiddenServiceManagementRoot(category) {
    return exports.HIDDEN_SERVICE_MANAGEMENT_CATEGORY_NAMES.has(normalizeCategoryName(category.name));
}
function filterServiceManagementCategoryTree(categories) {
    return categories.filter((category) => !isHiddenServiceManagementRoot(category));
}
exports.filterServiceManagementCategoryTree = filterServiceManagementCategoryTree;
function getHiddenServiceManagementCategoryIds(categories) {
    const ids = new Set();
    function collect(category) {
        ids.add(category.ID);
        (category.children || []).forEach(collect);
    }
    categories.forEach((category) => {
        if (isHiddenServiceManagementRoot(category)) {
            collect(category);
        }
    });
    return ids;
}
exports.getHiddenServiceManagementCategoryIds = getHiddenServiceManagementCategoryIds;
