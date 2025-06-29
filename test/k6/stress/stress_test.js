import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 1000,
    duration: '1s',
};

export default function () {
    const userId = Math.floor(Math.random() * 100000) + 1;
    const campaignId = 2;

    const payload = JSON.stringify({
        campaignId: campaignId,
        userId: userId,
    });

    const headers = {
        'Content-Type': 'application/json',
    };

    const res = http.post('http://localhost:8080/coupon.v1.CouponService/IssueCoupon', payload, { headers });

    check(res, {
        'status is 200': (r) => r.status === 200,
    });

    sleep(0.1);
}