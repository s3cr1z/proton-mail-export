include(cmake/mingw.cmake)
include(cmake/vcpkg_setup.cmake)
cmake_minimum_required(VERSION 3.23)
project(ExportTool CXX C)

set(CMAKE_MODULE_PATH ${CMAKE_CURRENT_SOURCE_DIR}/cmake)
if (UNIX AND NOT APPLE)
    set(CMAKE_INSTALL_RPATH "$\{ORIGIN\}")
endif()

enable_testing()

include(config)
include(compile_options)
include(clang_tidy)
include(clang_format)
include(shared_lib_copy)

# Find required packages for enhanced features
find_package(OpenSSL REQUIRED)
find_package(nlohmann_json CONFIG REQUIRED)

add_subdirectory(go-lib)
add_subdirectory(lib)
add_subdirectory(cli)
