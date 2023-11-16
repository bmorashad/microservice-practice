import http from 'k6/http'
import { k6SMLoadOptions } from './options.js'

export const options = k6SMLoadOptions

export default function() {
  let baseUrl = 'http://34.36.170.138'
  let response1 = http.get(`${baseUrl}/create-products/random`);
  let response2 = http.get(`${baseUrl}/products`);
  // console.log(response1.json());
  // console.log(response2.json());
}
