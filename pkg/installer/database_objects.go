package installer

var generatedDatabaseCRDV1 = `
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.5.0
  creationTimestamp: null
  name: databases.databases.schemahero.io
spec:
  group: databases.schemahero.io
  names:
    kind: Database
    listKind: DatabaseList
    plural: databases
    singular: database
  scope: Namespaced
  versions:
  - name: v1alpha4
    schema:
      openAPIV3Schema:
        description: Database is the Schema for the databases API
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            properties:
              connection:
                description: DatabaseConnection defines connection parameters for
                  the database driver
                properties:
                  cassandra:
                    properties:
                      hosts:
                        items:
                          type: string
                        type: array
                      keyspace:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                      password:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                      username:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                    required:
                    - hosts
                    - keyspace
                    type: object
                  cockroachdb:
                    properties:
                      dbname:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                      host:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                      password:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                      port:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                      sslmode:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                      uri:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                      user:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                    type: object
                  mysql:
                    properties:
                      collation:
                        type: string
                      dbname:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                      defaultCharset:
                        type: string
                      disableTLS:
                        type: boolean
                      host:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                      password:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                      port:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                      uri:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                      user:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                    type: object
                  postgres:
                    properties:
                      dbname:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                      host:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                      password:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                      port:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                      sslmode:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                      uri:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                      user:
                        properties:
                          value:
                            type: string
                          valueFrom:
                            properties:
                              secretKeyRef:
                                properties:
                                  key:
                                    type: string
                                  name:
                                    type: string
                                required:
                                - key
                                - name
                                type: object
                              ssm:
                                properties:
                                  accessKeyId:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  name:
                                    type: string
                                  region:
                                    type: string
                                  secretAccessKey:
                                    properties:
                                      value:
                                        type: string
                                      valueFrom:
                                        properties:
                                          secretKeyRef:
                                            properties:
                                              key:
                                                type: string
                                              name:
                                                type: string
                                            required:
                                            - key
                                            - name
                                            type: object
                                        type: object
                                    required:
                                    - value
                                    type: object
                                  withDecryption:
                                    type: boolean
                                required:
                                - name
                                type: object
                              vault:
                                properties:
                                  agentInject:
                                    type: boolean
                                  connectionTemplate:
                                    type: string
                                  endpoint:
                                    type: string
                                  kubernetesAuthEndpoint:
                                    type: string
                                  role:
                                    type: string
                                  secret:
                                    type: string
                                  serviceAccount:
                                    type: string
                                  serviceAccountNamespace:
                                    type: string
                                required:
                                - role
                                - secret
                                type: object
                            type: object
                        type: object
                    type: object
                  sqlite:
                    properties:
                      dsn:
                        type: string
                    required:
                    - dsn
                    type: object
                type: object
              enableShellCommand:
                type: boolean
              immediateDeploy:
                type: boolean
              schemahero:
                properties:
                  image:
                    type: string
                  nodeSelector:
                    additionalProperties:
                      type: string
                    type: object
                type: object
            type: object
          status:
            description: DatabaseStatus defines the observed state of Database
            properties:
              isConnected:
                type: boolean
              lastPing:
                type: string
            required:
            - isConnected
            - lastPing
            type: object
        required:
        - spec
        type: object
    served: true
    storage: true
    subresources:
      status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`
