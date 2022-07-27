import http from 'k6/http';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';
import { scenario } from 'k6/execution';
import { check } from 'k6';

let lastNumber = 0;

export const options = {
  scenarios: {
    create_profiles: {
      executor: 'ramping-arrival-rate',
      startRate: 10,
      preAllocatedVUs: 10,
      maxVUs: 50,
      stages: [
        {target: 10, duration: '20s'},
        {target: 40, duration: '20s'},
        {target: 100, duration: '20s'},
      ],
      exec: 'create_profile',
      startTime: '0s'
    },
    create_voucher: {
      executor: 'shared-iterations',
      vus: 10,
      iterations: 50,
      startTime: '1m',
      exec: 'create_voucher',
    },
    redeem_voucher: {
      executor: 'ramping-arrival-rate',
      startRate: 20,
      preAllocatedVUs: 10,
      maxVUs: 50,
      stages: [
        {target: 30, duration: '20s'},
        {target: 50, duration: '20s'},
        {target: 100, duration: '20s'},
      ],
      exec: 'redeem_voucher',
      startTime: '0s'
    },
  },
};

export function create_profile() {
  let body = JSON.stringify({
    phone_number: '+98913' + scenario.iterationInTest,
    balance: 0,
  });

  if (scenario.iterationInTest > lastNumber){
    lastNumber = scenario.iterationInTest;
  }

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const login_response = http.post('http://127.0.0.1:8001/api/profiles', body, params);
  check(login_response, {
    'is status 200': (r) => r.status === 200,
  });
}

console.log(lastNumber)

export function create_voucher() {
  let body = JSON.stringify({
    code: 'voucher_' + scenario.iterationInTest,
    amount: randomIntBetween(5000,10000),
    limit: randomIntBetween(100,1000),
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const login_response = http.post('http://127.0.0.1:8000/api/vouchers', body, params);
  check(login_response, {
    'is status 201': (r) => r.status === 201,
  });
}
export function redeem_voucher() {
  let body = JSON.stringify({
    phone_number: '+98913' + randomIntBetween(1, 2099),
    code: 'voucher_' + randomIntBetween(1, 50)
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const login_response = http.post('http://127.0.0.1:8000/api/vouchers/redeem', body, params);
  check(login_response, {
    'is status 200': (r) => r.status === 200,
  });
}
