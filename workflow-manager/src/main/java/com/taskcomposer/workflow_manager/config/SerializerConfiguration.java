package com.taskcomposer.workflow_manager.config;

import com.fasterxml.jackson.dataformat.yaml.YAMLMapper;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.MediaType;
import org.springframework.http.converter.HttpMessageConverter;
import org.springframework.http.converter.json.AbstractJackson2HttpMessageConverter;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

import java.util.List;

// https://stackoverflow.com/questions/37328314/spring-boot-restcontroller-deserializing-yaml-uploads
@Configuration
public class SerializerConfiguration implements WebMvcConfigurer {

    @Override
    public void extendMessageConverters(List<HttpMessageConverter<?>> converters) {
        converters.add(new YamlJackson2HttpMessageConverter());
    }

    static final class YamlJackson2HttpMessageConverter extends AbstractJackson2HttpMessageConverter {
        YamlJackson2HttpMessageConverter() {
            super(new YAMLMapper(),
                    MediaType.parseMediaType("application/yaml"),
                    MediaType.parseMediaType("application/*+yaml"),
                    MediaType.parseMediaType("application/yaml;charset=UTF-8"),
                    MediaType.parseMediaType("application/*+yaml;charset=UTF-8"));
        }
    }
}
