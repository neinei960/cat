"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.buildAppointmentCalendarSummary = void 0;
function buildAppointmentCalendarSummary(items) {
    return items.reduce((summary, item) => {
        const status = typeof item?.status === 'number' ? item.status : -1;
        const staffId = typeof item?.staff_id === 'number' ? item.staff_id : 0;
        if (status !== 4) {
            summary.total += 1;
        }
        if (status === 0) {
            summary.pendingConfirm += 1;
        }
        if (status !== 4 && !staffId) {
            summary.unassigned += 1;
        }
        if (status === 3) {
            summary.waitingCheckout += 1;
        }
        return summary;
    }, {
        total: 0,
        pendingConfirm: 0,
        unassigned: 0,
        waitingCheckout: 0,
    });
}
exports.buildAppointmentCalendarSummary = buildAppointmentCalendarSummary;
