package com.taskcomposer.service_registry.repositories.models;

import jakarta.persistence.*;
import java.util.List;

import java.util.Date;

@Entity
@Table(name = "services")
public class Service {

    @Id
    private String name;

    @Column(name = "broker_url")
    private String brokerUrl;

    @Column(name = "input_topic")
    private String inputTopic;

    @Column(name = "output_topic")
    private String outputTopic;

    @Column(name = "last_hearbeat")
    @Temporal(TemporalType.TIMESTAMP)
    private Date lastHeartbeat;

    @OneToMany(mappedBy = "service", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<Task> tasks;
}
