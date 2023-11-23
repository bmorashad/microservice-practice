import http from 'k6/http'
// import { k6SMLoadOptions } from './options.js'
import { apiEndpoints } from './api.js'
import { sleep } from 'k6';
// import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";
// import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.2/index.js';

export const options = {
  // linkerd wins without error
  scenarios: {
    // constant_request_rate: {
    //   executor: 'constant-arrival-rate',
    //   rate: 270,
    //   timeUnit: '1s', // 1000 iterations per second, i.e. 1000 RPS
    //   duration: '30s',
    //   // preAllocatedVUs: 100, // how large the initial pool of VUs would be
    //   preAllocatedVUs: 50, // how large the initial pool of VUs would be
    //   maxVUs: 200, // if the preAllocatedVUs are not enough, we can initialize more
    // },  
  // linkerd wins without error
    constant_request_rate: {
      executor: 'constant-arrival-rate',
      rate: 250,
      timeUnit: '1s', // 1000 iterations per second, i.e. 1000 RPS
      duration: '30s',
      // preAllocatedVUs: 100, // how large the initial pool of VUs would be
      preAllocatedVUs: 50, // how large the initial pool of VUs would be
      maxVUs: 150, // if the preAllocatedVUs are not enough, we can initialize more
    },  
  },
}

var baseUrl = `http://${__ENV.HOST}`;
var mesh = `${__ENV.MESH}` || `${__ENV.HOST}`

export function setup() {
  let r = http.get(`${baseUrl}/reset`);
  console.info("product reset result:", r.status);
}

export default function() {
  apiEndpoints.forEach(api => {
    let r = http.get(`${baseUrl}${api}`);
    sleep(Math.random() * 2)
    // console.log(r.json());
  });
  // let response1 = http.get(`${baseUrl}/create-products/random`);
  // let response2 = http.get(`${baseUrl}/products`);
  // console.log(response1.json());
  // console.log(response2.json());
}

// export function handleSummary(data) {
//   return {
//     `${mesh}-summary-multi.json`: JSON.stringify(data), //the default data object
//     `${mesh}-summary-multi.html`: htmlReport(data),
//     stdout: textSummary(data, { indent: '→', enableColors: true }),
//   };
// }

export function teardown(params) {
  let r = http.get(`${baseUrl}/products/count`);
  console.info("products count:", r.json())
  r = http.get(`${baseUrl}/reset`);
  console.info("product reset result:", r.status);
}
