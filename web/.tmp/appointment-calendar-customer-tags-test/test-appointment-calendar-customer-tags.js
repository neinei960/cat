"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const node_fs_1 = __importDefault(require("node:fs"));
const node_path_1 = __importDefault(require("node:path"));
const root = process.cwd();
const calendarPath = node_path_1.default.join(root, 'src/pages/appointment/calendar.vue');
const repoPath = node_path_1.default.resolve(root, '../server/internal/repository/appointment_repo.go');
const calendar = node_fs_1.default.readFileSync(calendarPath, 'utf8');
const repo = node_fs_1.default.readFileSync(repoPath, 'utf8');
function assert(condition, message) {
    if (!condition) {
        throw new Error(message);
    }
}
assert(repo.includes('Preload("Customer.CustomerTags"'), 'appointment repository should preload customer tags for calendar cards');
assert(calendar.includes('getCustomerTagItems(appt)'), 'calendar cards should derive owner/customer tag items');
assert(calendar.includes('tag-owner'), 'calendar should render owner tag chips with a distinct owner class');
assert(calendar.includes('ownerTags') && calendar.includes('petTags'), 'calendar should keep owner tags and pet tags in separate data buckets');
assert(calendar.includes('appt-owner-tag-row'), 'calendar should render owner tags on their own row');
assert(calendar.includes('.appt-tag.tag-owner'), 'calendar should style owner tags separately from pet tags');
console.log('appointment calendar customer tag checks passed');
