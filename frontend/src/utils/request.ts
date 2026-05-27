import axios from 'axios';
import { message } from 'antd';

const request = axios.create({
  baseURL: '/',
  timeout: 30000,
});

request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error),
);

request.interceptors.response.use(
  (response) => {
    const { data } = response;
    if (data.code !== 0) {
      message.error(data.message || 'request failed');
      return Promise.reject(new Error(data.message));
    }
    return data;
  },
  (error) => {
    if (error.response) {
      const { status } = error.response;
      if (status === 401) {
        localStorage.removeItem('token');
        window.location.href = '/login';
      }
      message.error(error.response.data?.message || 'network error');
    }
    return Promise.reject(error);
  },
);

export default request;
