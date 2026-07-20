"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const fs_1 = require("fs");
const path_1 = require("path");
const file = (0, fs_1.readFileSync)((0, path_1.resolve)(process.cwd(), 'src/pages/appointment/detail.vue'), 'utf8');
function assertContains(value, message) {
    if (!file.includes(value)) {
        throw new Error(message);
    }
}
assertContains('goCustomerDetail', 'appointment detail should expose customer detail navigation');
assertContains('/pages/customer/detail?id=', 'customer navigation should target customer detail');
assertContains('goPetDetail', 'appointment detail should expose pet detail navigation');
assertContains('/pages/pet/edit?id=', 'pet navigation should target pet edit/detail page');
assertContains('link-value', 'linked appointment fields should use the shared clickable style');
console.log('appointment detail navigation checks passed');
