"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const order_care_report_1 = require("../src/utils/order-care-report");
const web_image_save_1 = require("../src/utils/web-image-save");
function createSampleOrder(overrides = {}) {
    return {
        ID: 101,
        order_no: 'ORD-CARE-001',
        status: 1,
        pay_time: '2026-04-20 11:14:17',
        CreatedAt: '2026-04-19 09:00:00',
        customer: { ID: 1, shop_id: 1, openid: '', phone: '13800138000', nickname: '测试客户', avatar: '', gender: 0, remark: '', tags: '', total_spent: 0, visit_count: 0, last_visit_at: '', member_balance: 0, discount_rate: 1, address: '', address_detail: '', door_code: '', CreatedAt: '2026-04-19 09:00:00' },
        pet: {
            ID: 11,
            shop_id: 1,
            customer_id: 1,
            name: '福福',
            species: '猫',
            breed: '英短',
            gender: 1,
            birth_date: '2024-01-01',
            weight: 4.1,
            coat_type: '',
            coat_color: '',
            fur_level: '',
            personality: '',
            aggression: '',
            forbidden_zones: '',
            bath_frequency: '',
            neutered: true,
            avatar: '',
            care_notes: '',
            behavior_notes: '',
            status: 1,
            CreatedAt: '2026-01-01 00:00:00',
        },
        appointment: {
            pets: [
                {
                    pet_id: 11,
                    pet: {
                        ID: 11,
                        shop_id: 1,
                        customer_id: 1,
                        name: '福福',
                        species: '猫',
                        breed: '英短',
                        gender: 1,
                        birth_date: '2024-01-01',
                        weight: 4.1,
                        coat_type: '',
                        coat_color: '',
                        fur_level: '',
                        personality: '',
                        aggression: '',
                        forbidden_zones: '',
                        bath_frequency: '',
                        neutered: true,
                        avatar: '',
                        care_notes: '',
                        behavior_notes: '',
                        status: 1,
                        CreatedAt: '2026-01-01 00:00:00',
                    },
                },
                {
                    pet_id: 12,
                    pet: {
                        ID: 12,
                        shop_id: 1,
                        customer_id: 1,
                        name: '豆豆',
                        species: '猫',
                        breed: '布偶',
                        gender: 2,
                        birth_date: '2023-05-01',
                        weight: 5.2,
                        coat_type: '',
                        coat_color: '',
                        fur_level: '',
                        personality: '',
                        aggression: '',
                        forbidden_zones: '',
                        bath_frequency: '',
                        neutered: false,
                        avatar: '',
                        care_notes: '',
                        behavior_notes: '',
                        status: 1,
                        CreatedAt: '2026-01-01 00:00:00',
                    },
                },
            ],
        },
        pet_groups: [
            {
                pet_id: 11,
                pet_name: '福福',
                items: [
                    { ID: 1, order_id: 101, item_type: 1, item_id: 1001, name: '福福 · 皮毛调理', quantity: 1, unit_price: 88, amount: 88 },
                ],
            },
            {
                pet_id: 12,
                pet_name: '豆豆',
                items: [
                    { ID: 2, order_id: 101, item_type: 1, item_id: 1002, name: '福福 · Harmurry精致皮毛调理', quantity: 1, unit_price: 99, amount: 99 },
                ],
            },
        ],
        items: [
            { ID: 3, order_id: 101, item_type: 1, item_id: 1003, name: '通用 · 备用项目', quantity: 1, unit_price: 10, amount: 10 },
        ],
        ...overrides,
    };
}
function main() {
    const fixtureOrder = createSampleOrder();
    const unpaidOrder = createSampleOrder({ pay_status: 0 });
    const paidOrder = createSampleOrder({ pay_status: 1 });
    assertEqual((0, order_care_report_1.canGenerateOrderCareReport)(unpaidOrder), false, 'unpaid order cannot generate');
    assertEqual((0, order_care_report_1.canGenerateOrderCareReport)({ status: 0, pet_groups: [] }), false, 'unpaid empty order cannot generate');
    assertEqual((0, order_care_report_1.canGenerateOrderCareReport)(fixtureOrder), true, 'missing pay_status fixture can generate');
    assertEqual((0, order_care_report_1.canGenerateOrderCareReport)(paidOrder), true, 'paid order can generate');
    assertEqual((0, web_image_save_1.buildOrderCareReportFileName)('NO167', '福福'), '护理报告_NO167_福福.png', 'report filename');
    const petOptions = (0, order_care_report_1.buildOrderCareReportPetOptions)(fixtureOrder);
    assertEqual(petOptions.length, 2, 'pet options length');
    assertDeepEqual(petOptions.map((option) => option.petId), [11, 12], 'pet options ids');
    const draft = (0, order_care_report_1.buildOrderCareReportDraft)(fixtureOrder, 12);
    assertEqual(draft.petName, '豆豆', 'draft.petName');
    assertEqual(draft.breed, '布偶', 'draft.breed');
    assertEqual(draft.age, '2岁11月', 'draft.age');
    assertEqual(draft.careContent, 'Harmurry精致皮毛调理', 'draft.careContent');
    assertEqual((0, order_care_report_1.normalizeOrderCareReportDate)('2026.04.20'), '2026-04-20', 'normalize dotted date');
    assertEqual((0, order_care_report_1.normalizeOrderCareReportDate)('2026-04-25'), '2026-04-25', 'normalize dashed date');
    assertEqual(order_care_report_1.orderCareReportBodyShapeOptions.length, 5, 'body shape option count');
    assertEqual(order_care_report_1.orderCareReportSectionDefinitions.length, 7, 'section definition count');
    assertDeepEqual(order_care_report_1.orderCareReportSectionDefinitions.map((section) => section.key), ['skin', 'hair', 'nails', 'eyesFace', 'ears', 'oral', 'anus'], 'section keys');
}
main();
function assertEqual(actual, expected, label) {
    if (actual !== expected) {
        throw new Error(`${label}: expected ${JSON.stringify(expected)}, got ${JSON.stringify(actual)}`);
    }
}
function assertDeepEqual(actual, expected, label) {
    const actualJson = JSON.stringify(actual);
    const expectedJson = JSON.stringify(expected);
    if (actualJson !== expectedJson) {
        throw new Error(`${label}: expected ${expectedJson}, got ${actualJson}`);
    }
}
