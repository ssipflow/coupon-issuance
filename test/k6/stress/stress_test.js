import http from 'k6/http';
import { check } from 'k6';

export const options = {
    scenarios: {
        constant_rate_test: {
            executor: 'constant-arrival-rate',
            rate: 1000, // 1000 RPS
            timeUnit: '1s',
            duration: '1s',
            preAllocatedVUs: 1000,
            maxVUs: 2000,
        },
    },
};

export default function () {
    const userId = Math.floor(Math.random() * 100000) + 1;
    const campaignId = 4;

    const payload = JSON.stringify({
        campaignId: campaignId,
        userId: userId,
    });

    const headers = {
        'Content-Type': 'application/json',
    };

    const res = http.post('http://localhost:8080/coupon.v1.CouponService/IssueCoupon', payload, { headers });
    if (res.status !== 200) {
        const data = JSON.parse(res.body);
        console.log(`Failure issuing coupon for user ${userId}: ${data.message || 'Unknown error'}`);
    }

    check(res, {
        'status is 200': (r) => r.status === 200,
    });
}