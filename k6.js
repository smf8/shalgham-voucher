import http from 'k6/http';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';
import { scenario } from 'k6/execution';
import { check } from 'k6';

// let lastNumber = 0;

export const options = {
  scenarios: {
    create_profiles: {
      executor: 'ramping-arrival-rate',
      startRate: 10,
      preAllocatedVUs: 250,
      maxVUs: 1200,
      stages: [
        {target: 10, duration: '30s'},
        {target: 60, duration: '30s'},
        {target: 200, duration: '20s'},
        {target: 1000, duration: '60s'},
      ],
      exec: 'create_profile',
      startTime: '0s'
    },
    create_voucher: {
      executor: 'shared-iterations',
      vus: 50,
      iterations: 500,
      startTime: '0s',
      exec: 'create_voucher',
    },
    redeem_voucher: {
      executor: 'ramping-arrival-rate',
      startRate: 100,
      preAllocatedVUs: 4000,
      maxVUs: 5000,
      stages: [
        {target: 100, duration: '20s'},
        {target: 1500, duration: '10s'},
        {target: 300, duration: '10s'},
        {target: 3000, duration: '10s'},
        {target: 5000, duration: '10s'},
        {target: 100, duration: '30s'},
      ],
      exec: 'redeem_voucher',
      startTime: '30s'
    },
  },
};

export function create_profile() {
  let body = JSON.stringify({
    phone_number: '+98913' + (scenario.iterationInTest + 1000000),
    balance: 0,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const login_response = http.post('http://40.76.146.186:8000/api/profiles', body, params);
  check(login_response, {
    'is status 200': (r) => r.status === 200,
  });
}

export function create_voucher() {
  let body = JSON.stringify({
    code: 'voucher_' + scenario.iterationInTest,
    amount: randomIntBetween(5000,10000),
    limit: 2000,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const login_response = http.post('http://40.76.145.78:8000/api/vouchers', body, params);
  check(login_response, {
    'is status 201': (r) => r.status === 201,
  });
}
export function redeem_voucher() {
  let body = JSON.stringify({
    phone_number: '+98913' + randomIntBetween(1000000, 1299999),
    code: 'voucher_' + randomIntBetween(0, 499)
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const login_response = http.post('http://40.76.145.78:8000/api/vouchers/redeem', body, params);
  check(login_response, {
    'is status 200': (r) => r.status === 200,
  });
}
