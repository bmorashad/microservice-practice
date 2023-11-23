import http from 'k6/http'
import { k6SMLoadOptions } from './options.js'
import { apiEndpoints } from './api.js'
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.2/index.js';

export const options = k6SMLoadOptions

// var baseUrl = 'http://34.126.89.43:8010'
// var baseUrl = 'http://35.201.111.73'
var baseUrl = `http://${__ENV.HOST}`;
var mesh = `${__ENV.MESH}` || `${__ENV.HOST}`
export function setup() {
    let r = http.get(`${baseUrl}/reset`);
    console.info("product reset result:", r.status);
}

export default function() {
  apiEndpoints.forEach(api => {
    let r = http.get(`${baseUrl}${api}`);
    // console.log(r.json());
  });
  // let response1 = http.get(`${baseUrl}/create-products/random`);
  // let response2 = http.get(`${baseUrl}/products`);
  // console.log(response1.json());
  // console.log(response2.json());
}

export function handleSummary(data) {
  return {
    `${mesh}-summary-simple.json`: JSON.stringify(data), //the default data object
    `${mesh}-summary-simple.html`: htmlReport(data),
    stdout: textSummary(data, { indent: '→', enableColors: true }),
  };
}

export function teardown(params) {
  let r = http.get(`${baseUrl}/products/count`);
  console.info("products count:", r.json())
  r = http.get(`${baseUrl}/reset`);
  console.info("product reset result:", r.status);
}
