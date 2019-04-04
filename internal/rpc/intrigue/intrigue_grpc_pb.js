// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('grpc');
var intrigue_intrigue_pb = require('../intrigue/intrigue_pb.js');

function serialize_intrigue_Action(arg) {
  if (!(arg instanceof intrigue_intrigue_pb.Action)) {
    throw new Error('Expected argument of type intrigue.Action');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_intrigue_Action(buffer_arg) {
  return intrigue_intrigue_pb.Action.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_intrigue_DataRequest(arg) {
  if (!(arg instanceof intrigue_intrigue_pb.DataRequest)) {
    throw new Error('Expected argument of type intrigue.DataRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_intrigue_DataRequest(buffer_arg) {
  return intrigue_intrigue_pb.DataRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_intrigue_DataResponse(arg) {
  if (!(arg instanceof intrigue_intrigue_pb.DataResponse)) {
    throw new Error('Expected argument of type intrigue.DataResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_intrigue_DataResponse(buffer_arg) {
  return intrigue_intrigue_pb.DataResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_intrigue_EmptyRequest(arg) {
  if (!(arg instanceof intrigue_intrigue_pb.EmptyRequest)) {
    throw new Error('Expected argument of type intrigue.EmptyRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_intrigue_EmptyRequest(buffer_arg) {
  return intrigue_intrigue_pb.EmptyRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_intrigue_NewServiceRequest(arg) {
  if (!(arg instanceof intrigue_intrigue_pb.NewServiceRequest)) {
    throw new Error('Expected argument of type intrigue.NewServiceRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_intrigue_NewServiceRequest(buffer_arg) {
  return intrigue_intrigue_pb.NewServiceRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_intrigue_Ping(arg) {
  if (!(arg instanceof intrigue_intrigue_pb.Ping)) {
    throw new Error('Expected argument of type intrigue.Ping');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_intrigue_Ping(buffer_arg) {
  return intrigue_intrigue_pb.Ping.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_intrigue_Pong(arg) {
  if (!(arg instanceof intrigue_intrigue_pb.Pong)) {
    throw new Error('Expected argument of type intrigue.Pong');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_intrigue_Pong(buffer_arg) {
  return intrigue_intrigue_pb.Pong.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_intrigue_Receipt(arg) {
  if (!(arg instanceof intrigue_intrigue_pb.Receipt)) {
    throw new Error('Expected argument of type intrigue.Receipt');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_intrigue_Receipt(buffer_arg) {
  return intrigue_intrigue_pb.Receipt.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_intrigue_ServiceUpdate(arg) {
  if (!(arg instanceof intrigue_intrigue_pb.ServiceUpdate)) {
    throw new Error('Expected argument of type intrigue.ServiceUpdate');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_intrigue_ServiceUpdate(buffer_arg) {
  return intrigue_intrigue_pb.ServiceUpdate.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_intrigue_SummaryReceipt(arg) {
  if (!(arg instanceof intrigue_intrigue_pb.SummaryReceipt)) {
    throw new Error('Expected argument of type intrigue.SummaryReceipt');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_intrigue_SummaryReceipt(buffer_arg) {
  return intrigue_intrigue_pb.SummaryReceipt.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_intrigue_WhoIsRequest(arg) {
  if (!(arg instanceof intrigue_intrigue_pb.WhoIsRequest)) {
    throw new Error('Expected argument of type intrigue.WhoIsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_intrigue_WhoIsRequest(buffer_arg) {
  return intrigue_intrigue_pb.WhoIsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_intrigue_WhoIsResponse(arg) {
  if (!(arg instanceof intrigue_intrigue_pb.WhoIsResponse)) {
    throw new Error('Expected argument of type intrigue.WhoIsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_intrigue_WhoIsResponse(buffer_arg) {
  return intrigue_intrigue_pb.WhoIsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


var CabalService = exports.CabalService = {
  registerService: {
    path: '/intrigue.Cabal/RegisterService',
    requestStream: false,
    responseStream: false,
    requestType: intrigue_intrigue_pb.NewServiceRequest,
    responseType: intrigue_intrigue_pb.Receipt,
    requestSerialize: serialize_intrigue_NewServiceRequest,
    requestDeserialize: deserialize_intrigue_NewServiceRequest,
    responseSerialize: serialize_intrigue_Receipt,
    responseDeserialize: deserialize_intrigue_Receipt,
  },
  updateRegistration: {
    path: '/intrigue.Cabal/UpdateRegistration',
    requestStream: false,
    responseStream: false,
    requestType: intrigue_intrigue_pb.ServiceUpdate,
    responseType: intrigue_intrigue_pb.Receipt,
    requestSerialize: serialize_intrigue_ServiceUpdate,
    requestDeserialize: deserialize_intrigue_ServiceUpdate,
    responseSerialize: serialize_intrigue_Receipt,
    responseDeserialize: deserialize_intrigue_Receipt,
  },
  data: {
    path: '/intrigue.Cabal/Data',
    requestStream: false,
    responseStream: false,
    requestType: intrigue_intrigue_pb.DataRequest,
    responseType: intrigue_intrigue_pb.DataResponse,
    requestSerialize: serialize_intrigue_DataRequest,
    requestDeserialize: deserialize_intrigue_DataRequest,
    responseSerialize: serialize_intrigue_DataResponse,
    responseDeserialize: deserialize_intrigue_DataResponse,
  },
  whoIs: {
    path: '/intrigue.Cabal/WhoIs',
    requestStream: false,
    responseStream: false,
    requestType: intrigue_intrigue_pb.WhoIsRequest,
    responseType: intrigue_intrigue_pb.WhoIsResponse,
    requestSerialize: serialize_intrigue_WhoIsRequest,
    requestDeserialize: deserialize_intrigue_WhoIsRequest,
    responseSerialize: serialize_intrigue_WhoIsResponse,
    responseDeserialize: deserialize_intrigue_WhoIsResponse,
  },
  summary: {
    path: '/intrigue.Cabal/Summary',
    requestStream: false,
    responseStream: false,
    requestType: intrigue_intrigue_pb.Action,
    responseType: intrigue_intrigue_pb.SummaryReceipt,
    requestSerialize: serialize_intrigue_Action,
    requestDeserialize: deserialize_intrigue_Action,
    responseSerialize: serialize_intrigue_SummaryReceipt,
    responseDeserialize: deserialize_intrigue_SummaryReceipt,
  },
  alive: {
    path: '/intrigue.Cabal/Alive',
    requestStream: false,
    responseStream: false,
    requestType: intrigue_intrigue_pb.Ping,
    responseType: intrigue_intrigue_pb.Pong,
    requestSerialize: serialize_intrigue_Ping,
    requestDeserialize: deserialize_intrigue_Ping,
    responseSerialize: serialize_intrigue_Pong,
    responseDeserialize: deserialize_intrigue_Pong,
  },
};

exports.CabalClient = grpc.makeGenericClientConstructor(CabalService);
var RemoteService = exports.RemoteService = {
  notifyAction: {
    path: '/intrigue.Remote/NotifyAction',
    requestStream: false,
    responseStream: false,
    requestType: intrigue_intrigue_pb.Action,
    responseType: intrigue_intrigue_pb.Action,
    requestSerialize: serialize_intrigue_Action,
    requestDeserialize: deserialize_intrigue_Action,
    responseSerialize: serialize_intrigue_Action,
    responseDeserialize: deserialize_intrigue_Action,
  },
  summary: {
    path: '/intrigue.Remote/Summary',
    requestStream: false,
    responseStream: false,
    requestType: intrigue_intrigue_pb.Action,
    responseType: intrigue_intrigue_pb.SummaryReceipt,
    requestSerialize: serialize_intrigue_Action,
    requestDeserialize: deserialize_intrigue_Action,
    responseSerialize: serialize_intrigue_SummaryReceipt,
    responseDeserialize: deserialize_intrigue_SummaryReceipt,
  },
  updateRegistration: {
    path: '/intrigue.Remote/UpdateRegistration',
    requestStream: false,
    responseStream: false,
    requestType: intrigue_intrigue_pb.ServiceUpdate,
    responseType: intrigue_intrigue_pb.Receipt,
    requestSerialize: serialize_intrigue_ServiceUpdate,
    requestDeserialize: deserialize_intrigue_ServiceUpdate,
    responseSerialize: serialize_intrigue_Receipt,
    responseDeserialize: deserialize_intrigue_Receipt,
  },
  alive: {
    path: '/intrigue.Remote/Alive',
    requestStream: false,
    responseStream: false,
    requestType: intrigue_intrigue_pb.Ping,
    responseType: intrigue_intrigue_pb.Pong,
    requestSerialize: serialize_intrigue_Ping,
    requestDeserialize: deserialize_intrigue_Ping,
    responseSerialize: serialize_intrigue_Pong,
    responseDeserialize: deserialize_intrigue_Pong,
  },
};

exports.RemoteClient = grpc.makeGenericClientConstructor(RemoteService);
var ControlService = exports.ControlService = {
  startService: {
    path: '/intrigue.Control/StartService',
    requestStream: false,
    responseStream: false,
    requestType: intrigue_intrigue_pb.Action,
    responseType: intrigue_intrigue_pb.Receipt,
    requestSerialize: serialize_intrigue_Action,
    requestDeserialize: deserialize_intrigue_Action,
    responseSerialize: serialize_intrigue_Receipt,
    responseDeserialize: deserialize_intrigue_Receipt,
  },
  restartService: {
    path: '/intrigue.Control/RestartService',
    requestStream: false,
    responseStream: false,
    requestType: intrigue_intrigue_pb.Action,
    responseType: intrigue_intrigue_pb.Receipt,
    requestSerialize: serialize_intrigue_Action,
    requestDeserialize: deserialize_intrigue_Action,
    responseSerialize: serialize_intrigue_Receipt,
    responseDeserialize: deserialize_intrigue_Receipt,
  },
  killService: {
    path: '/intrigue.Control/KillService',
    requestStream: false,
    responseStream: false,
    requestType: intrigue_intrigue_pb.Action,
    responseType: intrigue_intrigue_pb.Receipt,
    requestSerialize: serialize_intrigue_Action,
    requestDeserialize: deserialize_intrigue_Action,
    responseSerialize: serialize_intrigue_Receipt,
    responseDeserialize: deserialize_intrigue_Receipt,
  },
  summary: {
    path: '/intrigue.Control/Summary',
    requestStream: false,
    responseStream: false,
    requestType: intrigue_intrigue_pb.Action,
    responseType: intrigue_intrigue_pb.SummaryReceipt,
    requestSerialize: serialize_intrigue_Action,
    requestDeserialize: deserialize_intrigue_Action,
    responseSerialize: serialize_intrigue_SummaryReceipt,
    responseDeserialize: deserialize_intrigue_SummaryReceipt,
  },
  updateRegistration: {
    path: '/intrigue.Control/UpdateRegistration',
    requestStream: false,
    responseStream: false,
    requestType: intrigue_intrigue_pb.ServiceUpdate,
    responseType: intrigue_intrigue_pb.Receipt,
    requestSerialize: serialize_intrigue_ServiceUpdate,
    requestDeserialize: deserialize_intrigue_ServiceUpdate,
    responseSerialize: serialize_intrigue_Receipt,
    responseDeserialize: deserialize_intrigue_Receipt,
  },
  alive: {
    path: '/intrigue.Control/Alive',
    requestStream: false,
    responseStream: false,
    requestType: intrigue_intrigue_pb.Ping,
    responseType: intrigue_intrigue_pb.Pong,
    requestSerialize: serialize_intrigue_Ping,
    requestDeserialize: deserialize_intrigue_Ping,
    responseSerialize: serialize_intrigue_Pong,
    responseDeserialize: deserialize_intrigue_Pong,
  },
  stopServer: {
    path: '/intrigue.Control/StopServer',
    requestStream: false,
    responseStream: false,
    requestType: intrigue_intrigue_pb.EmptyRequest,
    responseType: intrigue_intrigue_pb.Receipt,
    requestSerialize: serialize_intrigue_EmptyRequest,
    requestDeserialize: deserialize_intrigue_EmptyRequest,
    responseSerialize: serialize_intrigue_Receipt,
    responseDeserialize: deserialize_intrigue_Receipt,
  },
};

exports.ControlClient = grpc.makeGenericClientConstructor(ControlService);
