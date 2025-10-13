/*
 * Copyright 2025 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

const { expect, test, describe, beforeEach, afterEach } = require('@jest/globals');

// Helper to create a mock Apigee context
const mockContext = () => ({
  getVariable: jest.fn(),
  setVariable: jest.fn(),
});

// Array of test cases for each main script file we want to test.
const testCases = [
  {
    scriptPath: '../resources/jsc/filter-authorized-tools.cjs',
    functionName: 'filterAuthorizedTools',
    description: 'filter-authorized-tools.cjs main execution'
  },
  {
    scriptPath: '../resources/jsc/filter-header-tools.cjs',
    functionName: 'filterHeaderTools',
    description: 'filter-header-tools.cjs main execution'
  },
  {
    scriptPath: '../resources/jsc/process-rest-res.cjs',
    functionName: 'processRESTRes',
    description: 'process-rest-res.cjs main execution'
  },
  {
    scriptPath: '../resources/jsc/process-mcp-req.cjs',
    functionName: 'processMCPRequest',
    description: 'process-mcp-req.cjs main execution'
  },
  {
    scriptPath: '../resources/jsc/parse-mcp.cjs',
    functionName: 'parseMCPReq',
    description: 'parse-mcp.cjs main execution'
  }
];

// Use describe.each to run the same set of tests for each case defined above.
describe.each(testCases)('$description', ({ scriptPath, functionName }) => {

  beforeEach(() => {
    // Reset modules to ensure the script is re-executed for each test
    jest.resetModules();

    // Mock the global functions that the script expects to be available
    global[functionName] = jest.fn();
    global.setErrorResponse = jest.fn();
    global.isApigee = true; // Mock the Apigee environment flag
    global.print = jest.fn(); // Mock the Apigee print function
  });

  afterEach(() => {
    // Clean up globals to avoid polluting other tests
    delete global.context;
    delete global[functionName];
    delete global.setErrorResponse;
    delete global.isApigee;
    delete global.print;
  });

  test(`should call ${functionName} with the global context`, () => {
    const fakeContext = mockContext();
    global.context = fakeContext;

    require(scriptPath);

    expect(global[functionName]).toHaveBeenCalledTimes(1);
    expect(global[functionName]).toHaveBeenCalledWith(fakeContext);
    expect(global.setErrorResponse).not.toHaveBeenCalled();
  });

  test(`should call setErrorResponse if ${functionName} throws an error`, () => {
    const fakeContext = mockContext();
    global.context = fakeContext;

    const testError = new Error("Something went wrong");
    global[functionName].mockImplementation(() => {
      throw testError;
    });

    try {
      require(scriptPath);
    } catch (e) {
      // Expected because setErrorResponse throws.
    }

    expect(global[functionName]).toHaveBeenCalledWith(fakeContext);
    expect(global.setErrorResponse).toHaveBeenCalledTimes(1);
    expect(global.setErrorResponse).toHaveBeenCalledWith(fakeContext, 200, testError);
  });

  test('should call setErrorResponse with a ReferenceError if the function does not exist', () => {
    delete global[functionName]; // Make the function unavailable

    const fakeContext = mockContext();
    global.context = fakeContext;

    try {
      require(scriptPath);
    } catch (e) {
      // Expected.
    }

    expect(global.setErrorResponse).toHaveBeenCalledTimes(1);
    expect(global.setErrorResponse).toHaveBeenCalledWith(fakeContext, 200, expect.any(ReferenceError));
  });
});

