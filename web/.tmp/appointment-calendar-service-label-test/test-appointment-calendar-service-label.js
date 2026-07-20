"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const fs_1 = require("fs");
const path_1 = require("path");
const source = (0, fs_1.readFileSync)((0, path_1.resolve)(process.cwd(), 'src/pages/appointment/calendar.vue'), 'utf8');
function assertContains(text, message) {
    if (!source.includes(text)) {
        throw new Error(message);
    }
}
assertContains('function formatAppointmentServiceLabel', 'calendar should define a service label formatter');
assertContains('function isPricingRuleLabel', 'calendar should distinguish pricing-rule suffixes from add-on suffixes');
assertContains("baseName === '基础洗护' && ruleName === '超重'", 'formatter should handle overweight bath service labels');
assertContains('if (isPricingRuleLabel(ruleName)) return baseName', 'formatter should keep the root service label for pricing-rule suffixes like 伊珊娜·深层清洁护理(B)');
assertContains('return name', 'formatter should keep other service labels unchanged instead of showing only the suffix');
assertContains('.map((s: any) => formatAppointmentServiceLabel(s.service_name))', 'calendar pet service chips should use the formatter');
console.log('appointment calendar service label checks passed');
