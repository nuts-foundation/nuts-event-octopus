==================  ==========================  =========================================================================================================================
Key                 Default                     Description                                                                                                              
==================  ==========================  =========================================================================================================================
autoRecover         false                       Republish unfinished events at startup                                                                                   
connectionstring    file::memory:?cache=shared  db connection string for event store                                                                                     
incrementalBackoff  8                           Incremental backoff per retry queue, queue 0 retries after 1 second, queue 1 after {incrementalBackoff} * {previousDelay}
maxRetryCount       5                           Max number of retries for events before giving up (only for recoverable errors                                           
natsPort            4222                        Port for Nats to bind on                                                                                                 
purgeCompleted      false                       Purge completed events at startup                                                                                        
retryInterval       60                          Retry delay in seconds for reconnecting                                                                                  
==================  ==========================  =========================================================================================================================
