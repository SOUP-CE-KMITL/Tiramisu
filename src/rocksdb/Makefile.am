AUTOMAKE_OPTIONS = subdir-objects
OPT ?= -DNDEBUG

ACLOCAL_AMFLAGS = -I m4
WARNING_FLAGS = -Wall -Werror
ROCKSDB_CFLAGS = -g -O2 -fPIC
ROCKSDB_CXXFLAGS = -std=c++11 -g -O2 -fPIC
GFLAG = gflags
noinst_LTLIBRARIES = librocksdb.la

librocksdb_la_CFLAGS = $(WARNING_FLAGS) $(ROCKSDB_CFLAGS) -I$(srcdir) -I$(srcdir)/include -DOS_LINUX -DROCKSDB_PLATFORM_POSIX -fno-builtin-memcmp $(OPT) -DHAVE_JEMALLOC -Woverloaded-virtual
librocksdb_la_CXXFLAGS = $(WARNING_FLAGS) $(ROCKSDB_CXXFLAGS) -I$(srcdir) -I$(srcdir)/include -I$(srcdir)/third-party/gtest-1.7.0/fused-src -DOS_LINUX -DROCKSDB_PLATFORM_POSIX -fno-builtin-memcmp $(OPT) -DHAVE_JEMALLOC -Woverloaded-virtual -DROCKSDB_ATOMIC_PRESENT -DROCKSDB_FALLOCATE_PRESENT
librocksdb_la_LDFLAGS = -shared -Wl,-soname -Wl
if WITH_ATOMIC
   librocksdb_la_CFLAGS += -DROCKSDB_ATOMIC_PRESENT
   librocksdb_la_CXXFLAGS += -DROCKSDB_ATOMIC_PRESENT
endif

if WITH_FALLOCATE
   librocksdb_la_CFLAGS += -DROCKSDB_FALLOCATE_PRESENT
   librocksdb_la_CXXFLAGS += -DROCKSDB_FALLOCATE_PRESENT
endif
#
librocksdb_la_LDFLAGS += -lpthread
if WITH_TCMALLOC
   librocksdb_la_LDFLAGS += -ltcmalloc
if WITH_RT
   librocksdb_la_LDFLAGS += -lrt
endif
endif
if WITH_SNAPPY
   librocksdb_la_CFLAGS += -DSNAPPY
   librocksdb_la_CXXFLAGS += -DSNAPPY
   librocksdb_la_LDFLAGS += -lsnappy
endif
if WITH_Z
   librocksdb_la_CFLAGS += -DZLIB
   librocksdb_la_CXXFLAGS += -DZLIB
   librocksdb_la_LDFLAGS += -lz
endif
if WITH_BZ2
   librocksdb_la_CFLAGS += -DBZIP2
   librocksdb_la_CXXFLAGS += -DBZIP2
   librocksdb_la_LDFLAGS += -lbz2
endif
if WITH_GOOGLE_FLAGS
   GFLAGS = google
   librocksdb_la_CFLAGS += -DGFLAGS
   librocksdb_la_CXXFLAGS += -DGFLAGS
   librocksdb_la_LDFLAGS += -lgflags
endif
if WITH_GFLAGS_FLAGS
   GFLAGS = gflags
   librocksdb_la_CFLAGS += -DGFLAGS
   librocksdb_la_CXXFLAGS += -DGFLAGS
   librocksdb_la_LDFLAGS += -lgflags
endif

# Record the version of the source that we are compiling.
# We keep a record of the git revision in this file.  It is then built
# as a regular source file as part of the compilation process.
# One can run "strings executable_filename | grep _build_" to find
# the version of the source that we used to build the executable file.
date := $(shell date +%F)
git_sha := $(shell git describe HEAD 2>/dev/null)
gen_build_version =                                                     \
  printf '%s\n'                                                         \
    '\#include "build_version.h"'                                       \
    'const char* rocksdb_build_git_sha =                                \
      "rocksdb_build_git_sha:$(git_sha)";'                              \
    'const char* rocksdb_build_git_date =                               \
      "rocksdb_build_git_date:$(date)";'                                \
    'const char* rocksdb_build_compile_date = __DATE__;'
$(shell $(gen_build_version) > util/build_version.cc)

SOURCE_FILES = db/builder.cc \
        db/c.cc \
        db/column_family.cc \
        db/compaction.cc \
        db/compaction_job.cc \
        db/compaction_picker.cc \
        db/db_filesnapshot.cc \
        db/dbformat.cc \
        db/db_impl.cc \
        db/db_impl_debug.cc \
        db/db_impl_readonly.cc \
        db/db_iter.cc \
	db/event_logger_helpers.cc \
        db/file_indexer.cc \
        db/filename.cc \
        db/flush_job.cc \
        db/flush_scheduler.cc \
        db/forward_iterator.cc \
        db/internal_stats.cc \
        db/log_reader.cc \
        db/log_writer.cc \
        db/managed_iterator.cc \
        db/memtable_allocator.cc \
        db/memtable.cc \
        db/memtable_list.cc \
        db/merge_helper.cc \
        db/merge_operator.cc \
        db/repair.cc \
        db/slice.cc \
        db/table_cache.cc \
        db/table_properties_collector.cc \
        db/transaction_log_impl.cc \
        db/version_builder.cc \
        db/version_edit.cc \
        db/version_set.cc \
        db/wal_manager.cc \
        db/write_batch.cc \
        db/write_batch_base.cc \
        db/write_controller.cc \
        db/write_thread.cc \
        port/stack_trace.cc \
        port/port_posix.cc \
        table/adaptive_table_factory.cc \
        table/block_based_filter_block.cc \
        table/block_based_table_builder.cc \
        table/block_based_table_factory.cc \
        table/block_based_table_reader.cc \
        table/block_builder.cc \
        table/block.cc \
        table/block_hash_index.cc \
        table/block_prefix_index.cc \
        table/bloom_block.cc \
        table/cuckoo_table_builder.cc \
        table/cuckoo_table_factory.cc \
        table/cuckoo_table_reader.cc \
        table/flush_block_policy.cc \
        table/format.cc \
        table/full_filter_block.cc \
        table/get_context.cc \
        table/iterator.cc \
        table/merger.cc \
        table/meta_blocks.cc \
        table/plain_table_builder.cc \
        table/plain_table_factory.cc \
        table/plain_table_index.cc \
        table/plain_table_key_coding.cc \
        table/plain_table_reader.cc \
        table/table_properties.cc \
        table/two_level_iterator.cc \
        util/arena.cc \
        util/auto_roll_logger.cc \
        util/bloom.cc \
        util/build_version.cc \
        util/cache.cc \
        util/coding.cc \
        util/comparator.cc \
        util/crc32c.cc \
        util/db_info_dumper.cc \
        util/dynamic_bloom.cc \
        util/env.cc \
        util/env_hdfs.cc \
        util/env_posix.cc \
        util/file_util.cc \
        util/filter_policy.cc \
        util/hash.cc \
        util/hash_cuckoo_rep.cc \
        util/hash_linklist_rep.cc \
        util/hash_skiplist_rep.cc \
        util/histogram.cc \
        util/instrumented_mutex.cc \
        util/iostats_context.cc \
        utilities/backupable/backupable_db.cc \
        utilities/convenience/convenience.cc \
        utilities/checkpoint/checkpoint.cc \
        utilities/compacted_db/compacted_db_impl.cc \
        utilities/document/document_db.cc \
        utilities/document/json_document_builder.cc \
        utilities/document/json_document.cc \
        utilities/geodb/geodb_impl.cc \
        utilities/leveldb_options/leveldb_options.cc \
        utilities/merge_operators/put.cc \
        utilities/merge_operators/string_append/stringappend2.cc \
        utilities/merge_operators/string_append/stringappend.cc \
        utilities/merge_operators/uint64add.cc \
        utilities/redis/redis_lists.cc \
        utilities/spatialdb/spatial_db.cc \
        utilities/ttl/db_ttl_impl.cc \
        utilities/write_batch_with_index/write_batch_with_index.cc \
        util/event_logger.cc \
        util/ldb_cmd.cc \
        util/ldb_tool.cc \
        util/log_buffer.cc \
        util/logging.cc \
        util/memenv.cc \
        util/murmurhash.cc \
        util/mutable_cf_options.cc \
        util/options_builder.cc \
        util/options.cc \
        util/options_helper.cc \
        util/perf_context.cc \
        util/rate_limiter.cc \
        util/skiplistrep.cc \
        util/slice.cc \
        util/sst_dump_tool.cc \
        util/statistics.cc \
        util/status.cc \
        util/string_util.cc \
        util/sync_point.cc \
        util/thread_local.cc \
        util/thread_status_impl.cc \
        util/thread_status_updater.cc \
        util/thread_status_updater_debug.cc \
        util/thread_status_util.cc \
        util/thread_status_util_debug.cc \
        util/vectorrep.cc \
        util/xfunc.cc \
        util/xxhash.cc

SOURCE_H = util/allocator.h \
        util/arena.h \
        util/auto_roll_logger.h \
        util/autovector.h \
        util/build_version.h \
        util/coding.h \
        util/compression.h \
        util/crc32c.h \
        util/db_info_dumper.h \
        util/dynamic_bloom.h \
        util/event_logger.h \
        util/file_util.h \
        util/hash.h \
        util/hash_cuckoo_rep.h \
        util/hash_linklist_rep.h \
        util/hash_skiplist_rep.h \
        util/histogram.h \
        util/instrumented_mutex.h \
        util/iostats_context_imp.h \
        util/ldb_cmd.h \
        util/ldb_cmd_execute_result.h \
        util/log_buffer.h \
        util/logging.h \
        util/mock_env.h \
        util/murmurhash.h \
        util/mutable_cf_options.h \
        util/mutexlock.h \
        util/options_helper.h \
        util/perf_context_imp.h \
        util/posix_logger.h \
        util/random.h \
        util/rate_limiter.h \
        util/scoped_arena_iterator.h \
        util/sst_dump_tool_imp.h \
        util/statistics.h \
        util/stl_wrappers.h \
        util/stop_watch.h \
        util/string_util.h \
        util/sync_point.h \
        util/testharness.h \
        util/testutil.h \
        util/thread_local.h \
        util/thread_operation.h \
        util/thread_status_updater.h \
        util/thread_status_util.h \
        util/xfunc.h \
        util/xxhash.h \
        db/builder.h \
        db/column_family.h \
        db/compaction.h \
        db/compaction_job.h \
        db/compaction_picker.h \
        db/db_impl.h \
        db/db_impl_readonly.h \
        db/db_iter.h \
        db/dbformat.h \
        db/file_indexer.h \
        db/filename.h \
        db/flush_job.h \
        db/flush_scheduler.h \
        db/forward_iterator.h \
        db/internal_stats.h \
        db/job_context.h \
        db/log_format.h \
        db/log_reader.h \
        db/log_writer.h \
        db/managed_iterator.h \
        db/memtable.h \
        db/memtable_allocator.h \
        db/memtable_list.h \
        db/merge_context.h \
        db/merge_helper.h \
        db/skiplist.h \
        db/snapshot.h \
        db/table_cache.h \
        db/table_properties_collector.h \
        db/transaction_log_impl.h \
        db/version_builder.h \
        db/version_edit.h \
        db/version_set.h \
        db/wal_manager.h \
        db/write_batch_internal.h \
        db/write_controller.h \
        db/write_thread.h \
        db/writebuffer.h \
        table/adaptive_table_factory.h \
        table/block.h \
        table/block_based_filter_block.h \
        table/block_based_table_builder.h \
        table/block_based_table_factory.h \
        table/block_based_table_reader.h \
        table/block_builder.h \
        table/block_hash_index.h \
        table/block_prefix_index.h \
        table/bloom_block.h \
        table/cuckoo_table_builder.h \
        table/cuckoo_table_factory.h \
        table/cuckoo_table_reader.h \
        table/filter_block.h \
        table/format.h \
        table/full_filter_block.h \
        table/get_context.h \
        table/iter_heap.h \
        table/iterator_wrapper.h \
        table/merger.h \
        table/meta_blocks.h \
        table/mock_table.h \
        table/plain_table_builder.h \
        table/plain_table_factory.h \
        table/plain_table_index.h \
        table/plain_table_key_coding.h \
        table/plain_table_reader.h \
        table/table_builder.h \
        table/table_properties_internal.h \
        table/table_reader.h \
        table/two_level_iterator.h \
	include/utilities/geo_db.h \
	include/utilities/stackable_db.h \
	include/utilities/db_ttl.h \
	include/utilities/utility_db.h \
	include/utilities/backupable_db.h \
        include/rocksdb/c.h \
        include/rocksdb/cache.h \
        include/rocksdb/compaction_filter.h \
        include/rocksdb/comparator.h \
        include/rocksdb/db.h \
        include/rocksdb/env.h \
        include/rocksdb/filter_policy.h \
        include/rocksdb/flush_block_policy.h \
        include/rocksdb/immutable_options.h \
        include/rocksdb/iostats_context.h \
        include/rocksdb/iterator.h \
        include/rocksdb/ldb_tool.h \
        include/rocksdb/listener.h \
        include/rocksdb/memtablerep.h \
        include/rocksdb/merge_operator.h \
        include/rocksdb/metadata.h \
        include/rocksdb/options.h \
        include/rocksdb/perf_context.h \
        include/rocksdb/rate_limiter.h \
        include/rocksdb/slice.h \
        include/rocksdb/slice_transform.h \
        include/rocksdb/sst_dump_tool.h \
        include/rocksdb/statistics.h \
        include/rocksdb/status.h \
        include/rocksdb/table.h \
        include/rocksdb/table_properties.h \
        include/rocksdb/thread_status.h \
        include/rocksdb/transaction_log.h \
        include/rocksdb/types.h \
        include/rocksdb/universal_compaction.h \
        include/rocksdb/version.h \
        include/rocksdb/write_batch.h \
        include/rocksdb/write_batch_base.h \
        include/rocksdb/utilities/backupable_db.h \
        include/rocksdb/utilities/checkpoint.h \
        include/rocksdb/utilities/convenience.h \
        include/rocksdb/utilities/db_ttl.h \
        include/rocksdb/utilities/document_db.h \
        include/rocksdb/utilities/geo_db.h \
        include/rocksdb/utilities/json_document.h \
        include/rocksdb/utilities/leveldb_options.h \
        include/rocksdb/utilities/spatial_db.h \
        include/rocksdb/utilities/stackable_db.h \
        include/rocksdb/utilities/utility_db.h \
        include/rocksdb/utilities/write_batch_with_index.h

noinst_HEADERS = $(SOURCE_H)

librocksdb_la_SOURCES = $(SOURCE_FILES)
