import http from "k6/http";
// import { k6SMLoadOptions } from "./options.js";
import { apiEndpoints } from "./api.js";
import { sleep } from "k6";
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";
import { textSummary } from "https://jslib.k6.io/k6-summary/0.0.2/index.js";

const k6Options = {
  scenarios: {
    // constant_request_rate: { // Test Sample
    //   executor: "constant-arrival-rate",
    //   rate: 1,
    //   timeUnit: "1s",
    //   duration: "3s",
    //   preAllocatedVUs: 1, // how large the initial pool of VUs would be
    //   startTime: "0s",
    // },
    constant_request_rate_250: {
      // linkerd wins without error
      executor: "constant-arrival-rate",
      rate: 150,
      timeUnit: "1s", // 1000 iterations per second, i.e. 1000 RPS
      duration: "3m",
      // preAllocatedVUs: 100, // how large the initial pool of VUs would be
      preAllocatedVUs: 50, // how large the initial pool of VUs would be
      maxVUs: 200, // if the preAllocatedVUs are not enough, we can initialize more
      startTime: "0s",
    },
    constant_request_rate_230: {
      // linkerd wins without error
      executor: "constant-arrival-rate",
      rate: 100,
      timeUnit: "1s", // 1000 iterations per second, i.e. 1000 RPS
      duration: "2m",
      // preAllocatedVUs: 100, // how large the initial pool of VUs would be
      preAllocatedVUs: 50, // how large the initial pool of VUs would be
      maxVUs: 150, // if the preAllocatedVUs are not enough, we can initialize more
      startTime: "3.2m",
    },
    ramping_vus: {
      // istio wins by small margin
      executor: "ramping-vus",
      startvus: 50,
      stages: [
        { duration: "1m", target: 150 },
        { duration: "1m", target: 100 },
        { duration: "1m", target: 50 },
        { duration: "30s", target: 0 },
      ],
      gracefulRampDown: "1s",
      startTime: "5.2m",
    },
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
    //   startTime: "2m",
    // },
    // ramping_arrival_rate: {
    //   //Not Working Properly
    //   executor: "ramping-arrival-rate",
    //   startRate: 100,
    //   timeUnit: "1m",
    //   preAllocatedVUs: 25,
    //   stages: [
    //     { target: 50, duration: "1m" },
    //     { target: 100, duration: "1.5m" },
    //     { target: 150, duration: "1.5m" },
    //     { target: 10, duration: "1m" },
    //     { target: 0, duration: "5s" },
    //   ],
    //   startTime: "2m",
    // },
  },
};

let opts = {};

export const options = Object.assign(opts, k6Options);

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
  const day = String(now.getDate()).padStart(2, "0");
  const month = String(now.getMonth() + 1).padStart(2, "0"); // Month is zero-based
  const year = String(now.getFullYear()).slice(-2); // Get last two digits of the year
  const hours = String(now.getHours()).padStart(2, "0");
  const minutes = String(now.getMinutes()).padStart(2, "0");
  // Construct the formatted date-time string
  const formattedDateTime = `${day}-${month}-${year}-${hours}:${minutes}`;
  return formattedDateTime;
}

export function handleSummary(data) {
  const now = getNowDate();
  const fileName = `${mesh}-results/${mesh}-summary--epoch:${now}`;
  const fileNameJson = `${fileName}.json`;
  const fileNameHtml = `${fileName}.html`;
  const fileNameTxt = `${fileName}.txt`;
  data["testScenarios"] = k6Options.scenarios;
  const summaryHandlerOpts = {};
  summaryHandlerOpts[fileNameJson] = JSON.stringify(data);
  summaryHandlerOpts[fileNameHtml] = htmlReport(data);
  summaryHandlerOpts[fileNameTxt] = textSummary(data, { enableColors: true });
  summaryHandlerOpts["stdout"] = textSummary(data, {
    indent: "→",
    enableColors: true,
  });

  return summaryHandlerOpts;
}

export function teardown(_) {
  let r = http.get(`${baseUrl}/products/count`);
  console.info("products count:", r.json());
  r = http.get(`${baseUrl}/reset`);
  console.info("product reset result:", r.status);
}
