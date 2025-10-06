# TODO

- [ ] Add handleMockWebhook function to reduce cognitive complexity
- [ ] Modify /merchant-webhook route to use handleMockWebhook
- [ ] Replace if (process.env.NODE_ENV !== 'production') with optional chaining
- [ ] Fix error status codes in create-merchant-payment-intent endpoint (return 400 for validation errors)
- [ ] Improve webhook handling in merchant-webhook endpoint
- [ ] Test fixes with comprehensive merchant test suite
- [ ] Verify all endpoints work correctly in both mock and real modes
