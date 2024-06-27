plugins {
  `java-library`
}

repositories {
  mavenCentral()
}

dependencies {
  implementation("org.springframework.security:spring-security-crypto:5.7.3")
}

dependencyLocking {
  lockAllConfigurations()
}
