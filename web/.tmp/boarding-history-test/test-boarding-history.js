"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const boarding_history_1 = require("../src/utils/boarding-history");
function assertEqual(actual, expected, label) {
    if (actual !== expected) {
        throw new Error(`${label}: expected "${String(expected)}", got "${String(actual)}"`);
    }
}
function assertDeepEqual(actual, expected, label) {
    const actualJson = JSON.stringify(actual);
    const expectedJson = JSON.stringify(expected);
    if (actualJson !== expectedJson) {
        throw new Error(`${label}: expected ${expectedJson}, got ${actualJson}`);
    }
}
function createOrderFixture() {
    return {
        remark: '  猫粮分早晚两次，晚上需要单独安抚一下情绪。  ',
        customer: {
            nickname: 'happy',
            phone: '15900000000',
            remark: '对陌生人慢热',
        },
        cabinet: {
            cabinet_type: '唐娜温柔乡',
        },
        rooms: [
            {
                room_index: 1,
                cabinet: { cabinet_type: '唐娜温柔乡' },
            },
            {
                room_index: 2,
                cabinet: { cabinet_type: '波妞的游乐园' },
            },
        ],
        pets: [
            {
                pet_name_snapshot: '收养小猫',
                pet: {
                    name: '收养小猫',
                    breed: '银渐层',
                    gender: 1,
                    birth_date: '2024-01-01',
                },
            },
            {
                pet_name_snapshot: '月光',
                pet: {
                    name: '月光',
                    breed: '德文',
                    gender: 2,
                    birth_date: '',
                },
            },
        ],
    };
}
function testCustomerLabelPrefersNickname() {
    const actual = (0, boarding_history_1.getBoardingHistoryCustomerLabel)(createOrderFixture());
    assertEqual(actual, 'happy', 'customer.nickname');
}
function testRoomSummaryUsesMultiRoomCopy() {
    const actual = (0, boarding_history_1.getBoardingHistoryRoomSummary)(createOrderFixture());
    assertEqual(actual, '2 个房间 · 唐娜温柔乡、波妞的游乐园', 'room.summary.multi');
}
function testRemarkSummaryTruncatesCleanly() {
    const actual = (0, boarding_history_1.getBoardingHistoryRemarkSummary)(createOrderFixture().remark, 14);
    assertEqual(actual, '猫粮分早晚两次，晚上需要单…', 'remark.summary');
}
function testBuildPetProfilesIncludesBreedGenderAndAge() {
    const actual = (0, boarding_history_1.buildBoardingHistoryPetProfiles)(createOrderFixture());
    assertDeepEqual(actual, [
        { name: '收养小猫', meta: '银渐层 · 公 · 2岁3个月' },
        { name: '月光', meta: '德文 · 母' },
    ], 'pet.profiles');
}
function main() {
    testCustomerLabelPrefersNickname();
    testRoomSummaryUsesMultiRoomCopy();
    testRemarkSummaryTruncatesCleanly();
    testBuildPetProfilesIncludesBreedGenderAndAge();
    console.log('boarding-history tests passed');
}
main();
