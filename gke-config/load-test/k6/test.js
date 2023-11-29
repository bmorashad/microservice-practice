import http from "k6/http";
// import { k6SMLoadOptions } from "./options.js";
import { apiEndpoints } from "./api.js";
import { sleep } from "k6";
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.2/index.js';

export const options = {
  scenarios: {
    constant_request_rate: { // linkerd wins without error
      executor: "constant-arrival-rate",
      rate: 250,
      timeUnit: "1s", // 1000 iterations per second, i.e. 1000 RPS
      // duration: "30s",
      duration: "5s",
      // preAllocatedVUs: 100, // how large the initial pool of VUs would be
      preAllocatedVUs: 50, // how large the initial pool of VUs would be
      maxVUs: 200, // if the preAllocatedVUs are not enough, we can initialize more
    },
    // constant_request_rate: { // linkerd wins without error
    //   executor: "constant-arrival-rate",
    //   rate: 250,
    //   timeUnit: "1s", // 1000 iterations per second, i.e. 1000 RPS
    //   duration: "30s",
    //   // preAllocatedVUs: 100, // how large the initial pool of VUs would be
    //   preAllocatedVUs: 50, // how large the initial pool of VUs would be
    //   maxVUs: 150, // if the preAllocatedVUs are not enough, we can initialize more
    // },
    // ramping_vus: {
    //   // istio wins by small margin
    //   executor: "ramping-vus",
    //   startvus: 15,
    //   stages: [
    //     { duration: "30s", target: 150 },
    //     { duration: "10s", target: 100 },
    //     { duration: "10s", target: 50 },
    //     { duration: "5s", target: 0 },
    //   ],
    //   gracefulRampDown: "1s",
    //   // startTime: "10s",
    // },
    // ramping_vus: {
    //   executor: "ramping-vus",
    //   startvus: 25,
    //   stages: [
    //     // { duration: "30s", target: 250 }, // test again
    //     // { duration: "30s", target: 350 }, // linkerd wins no failiure
    //     // { duration: "30s", target: 300 }, // linkerd wins no failiure
    //     { duration: "30s", target: 280 }, // linkerd overall slightly faster, looses in p95
    //     { duration: "20s", target: 100 },
    //     { duration: "10s", target: 50 },
    //     { duration: "5s", target: 0 },
    //   ],
    //   gracefulRampDown: "1s",
    //   // startTime: "10s",
    // },
    // Not Working Properly
    // ramping_arrival_rate: {
    //   executor: 'ramping-arrival-rate',
    //   startRate: 100,
    //   timeUnit: '1m',
    //   preAllocatedVUs: 25,
    //   stages: [
    //     { target: 100, duration: '1m' },
    //     { target: 150, duration: '1.5m' },
    //     { target: 250, duration: '1.5m' },
    //     { target: 10, duration: '1m' },
    //     { target: 0, duration: '5s' },
    //   ],
    //   // startTime: "20s",
    // },
  },
};

var baseUrl = `http://${__ENV.HOST}`;
var mesh = `${__ENV.MESH}` || `${__ENV.HOST}`;

export function setup() {
  let r = http.get(`${baseUrl}/reset`);
  console.info("product reset result:", r.status);
}

export default function () {
  apiEndpoints.forEach((api) => {
    let r = http.get(`${baseUrl}${api}`);
    sleep(Math.random() * 2);
    // console.log(r.json());
  });
  // let response1 = http.get(`${baseUrl}/create-products/random`);
  // let response2 = http.get(`${baseUrl}/products`);
  // console.log(response1.json());
  // console.log(response2.json());
}

function getNowDate() {
  const now = new Date();
  // Get individual date and time components
  const day = String(now.getDate()).padStart(2, '0');
  const month = String(now.getMonth() + 1).padStart(2, '0'); // Month is zero-based
  const year = String(now.getFullYear()).slice(-2); // Get last two digits of the year
  const hours = String(now.getHours()).padStart(2, '0');
  const minutes = String(now.getMinutes()).padStart(2, '0');
  // Construct the formatted date-time string
  const formattedDateTime = `${day}-${month}-${year}-${hours}:${minutes}`;
  return formattedDateTime;
}

export function handleSummary(data) {
  const now = getNowDate()
  const fileName = `${mesh}-results/${mesh}-multi-test-results--epoch:${now}`
  const fileNameJson = `${fileName}.json`
  const fileNameHtml = `${fileName}.html`
  const fileNameSummary = `${mesh}-results/${mesh}-multi-test-summary--epoch:${now}.txt`
  data.testOpts = options.scenarios
  const summaryHandlerOpts = {}
  summaryHandlerOpts[fileNameJson] = JSON.stringify(data)
  summaryHandlerOpts[fileNameHtml] = htmlReport(data)
  summaryHandlerOpts[fileNameSummary]= textSummary(data, {enableColors: true})
  summaryHandlerOpts['stdout']= textSummary(data, { indent: '→', enableColors: true })

  return summaryHandlerOpts
}

export function teardown(_) {
  let r = http.get(`${baseUrl}/products/count`);
  console.info("products count:", r.json());
  r = http.get(`${baseUrl}/reset`);
  console.info("product reset result:", r.status);
}