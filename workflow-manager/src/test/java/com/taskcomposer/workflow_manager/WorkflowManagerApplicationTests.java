package com.taskcomposer.workflow_manager;

import com.taskcomposer.workflow_manager.config.ServiceConfig;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.context.annotation.Import;

@Tag("integration")
@Import(TestcontainersConfiguration.class)
@SpringBootTest
class WorkflowManagerApplicationTests {

	@Test
	void contextLoads() {
	}

}
