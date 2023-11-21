import http from 'k6/http'
import { k6SMLoadOptions } from './options.js'
import { apiEndpoints } from './api.js'
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.2/index.js';

export const options = k6SMLoadOptions

export default function() {
  let baseUrl = 'http://34.126.78.156:8010'
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
    'linkerd-summary.json': JSON.stringify(data), //the default data object
    "linkerd-summary.html": htmlReport(data),
    stdout: textSummary(data, { indent: '→', enableColors: true }),
  };
}